package relationships

import (
	"context"
	"fmt"
	"time"

	"github.com/uptrace/bun"

	"github.com/nhymxu/kith-pms/internal/people"
)

// GraphNode is a person vertex in the relationship network.
type GraphNode struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	Nickname    string   `json:"nickname"`
	Avatar      string   `json:"avatar"`
	Group       string   `json:"group"`  // first label — used for canvas colour
	Groups      []string `json:"groups"` // all labels — used for display
	IsSelf      bool     `json:"is_self"`
	DateOfBirth *string  `json:"date_of_birth,omitempty"`
	LastContact *string  `json:"last_contact_at,omitempty"`
}

// GraphLink is a deduplicated edge between two people.
// Source is always the lower ID; Target is the higher ID.
type GraphLink struct {
	Source      int64  `json:"source"`
	Target      int64  `json:"target"`
	Type        string `json:"type"`
	ReverseType string `json:"reverse_type"`
}

// Graph is the full response payload for the graph endpoint.
type Graph struct {
	Nodes []GraphNode `json:"nodes"`
	Links []GraphLink `json:"links"`
}

// graphEdgeRow is a raw row returned by GraphEdges query.
type graphEdgeRow struct {
	FromPersonID    int64      `bun:"from_person_id"`
	ToPersonID      int64      `bun:"to_person_id"`
	FromName        string     `bun:"from_name"`
	FromNickname    string     `bun:"from_nickname"`
	FromAvatar      string     `bun:"from_avatar"`
	FromIsSelf      bool       `bun:"from_is_self"`
	FromDateOfBirth *string    `bun:"from_date_of_birth"`
	FromLastContact *time.Time `bun:"from_last_contact_at"`
	ToName          string     `bun:"to_name"`
	ToNickname      string     `bun:"to_nickname"`
	ToAvatar        string     `bun:"to_avatar"`
	ToIsSelf        bool       `bun:"to_is_self"`
	ToDateOfBirth   *string    `bun:"to_date_of_birth"`
	ToLastContact   *time.Time `bun:"to_last_contact_at"`
	TypeName        string     `bun:"type_name"`
	ReverseName     string     `bun:"reverse_name"`
}

// GraphEdgesRepo is the read-only repo method for fetching graph edge data.
type GraphEdgesRepo interface {
	GraphEdges(ctx context.Context, personID int64) ([]graphEdgeRow, error)
}

type sqlGraphEdgesRepo struct{ db *bun.DB }

func newSQLGraphEdgesRepo(db *bun.DB) GraphEdgesRepo {
	return &sqlGraphEdgesRepo{db: db}
}

// GraphEdges fetches all junction rows with person + type metadata.
// When personID > 0, restricts to the ego-network (focal person + all neighbors
// + edges among neighbors). When personID == 0, returns the full global graph.
func (r *sqlGraphEdgesRepo) GraphEdges(ctx context.Context, personID int64) ([]graphEdgeRow, error) {
	const baseQuery = `` + //nolint:lll
		`SELECT pr.from_person_id, pr.to_person_id,` +
		` pf.name AS from_name, COALESCE(pf.nickname,'') AS from_nickname,` +
		` COALESCE(pf.avatar_path,'') AS from_avatar, pf.is_self AS from_is_self,` +
		` pf.date_of_birth AS from_date_of_birth, pf.last_contact_at AS from_last_contact_at,` +
		` pt.name AS to_name, COALESCE(pt.nickname,'') AS to_nickname,` +
		` COALESCE(pt.avatar_path,'') AS to_avatar, pt.is_self AS to_is_self,` +
		` pt.date_of_birth AS to_date_of_birth, pt.last_contact_at AS to_last_contact_at,` +
		` rt.name AS type_name, rt.reverse_name` +
		` FROM person_relationship pr` +
		` JOIN relationship_type rt ON rt.id = pr.relationship_type_id` +
		` JOIN person pf ON pf.id = pr.from_person_id` +
		` JOIN person pt ON pt.id = pr.to_person_id`

	if personID == 0 {
		var rows []graphEdgeRow
		if err := r.db.NewRaw(baseQuery).Scan(ctx, &rows); err != nil {
			return nil, fmt.Errorf("relationships: graph edges global: %w", err)
		}

		return rows, nil
	}

	// Ego mode: collect included node IDs = focal + all direct neighbors.
	type pairRow struct {
		From int64 `bun:"from_person_id"`
		To   int64 `bun:"to_person_id"`
	}

	var pairs []pairRow

	err := r.db.NewRaw(
		`SELECT from_person_id, to_person_id FROM person_relationship WHERE from_person_id = ? OR to_person_id = ?`,
		personID, personID,
	).Scan(ctx, &pairs)
	if err != nil {
		return nil, fmt.Errorf("relationships: graph edges ego neighbors: %w", err)
	}

	included := make(map[int64]struct{})
	included[personID] = struct{}{}

	for _, p := range pairs {
		included[p.From] = struct{}{}
		included[p.To] = struct{}{}
	}

	ids := make([]int64, 0, len(included))
	for id := range included {
		ids = append(ids, id)
	}

	// Return all rows where BOTH endpoints are in the included set.
	var rows []graphEdgeRow

	err = r.db.NewRaw(
		baseQuery+` WHERE pr.from_person_id IN (?) AND pr.to_person_id IN (?)`,
		bun.List(ids), bun.List(ids),
	).Scan(ctx, &rows)
	if err != nil {
		return nil, fmt.Errorf("relationships: graph edges ego: %w", err)
	}

	return rows, nil
}

// avatarURL returns the canonical avatar URL for the person ID when avatarPath is non-empty.
func avatarURL(id int64, avatarPath string) string {
	if avatarPath == "" {
		return ""
	}

	return fmt.Sprintf("/v1/people/%d/avatar", id)
}

// timeToDateStr converts a nullable time.Time to a nullable ISO-8601 date string.
func timeToDateStr(t *time.Time) *string {
	if t == nil {
		return nil
	}

	s := t.UTC().Format(time.RFC3339)

	return &s
}

// Graph builds the deduplicated node+link graph.
// personID == 0 → global graph; personID > 0 → ego-network.
func (s *Service) Graph(ctx context.Context, personID int64) (Graph, error) {
	rows, err := s.graphEdges.GraphEdges(ctx, personID)
	if err != nil {
		return Graph{}, err
	}

	type nodeData struct {
		name        string
		nickname    string
		avatar      string
		isSelf      bool
		dateOfBirth *string
		lastContact *string
	}

	nodeMap := make(map[int64]nodeData)
	linkMap := make(map[string]GraphLink)

	for _, row := range rows {
		// Populate node map from each row's from/to person data.
		if _, ok := nodeMap[row.FromPersonID]; !ok {
			nodeMap[row.FromPersonID] = nodeData{
				name:        row.FromName,
				nickname:    row.FromNickname,
				avatar:      row.FromAvatar,
				isSelf:      row.FromIsSelf,
				dateOfBirth: row.FromDateOfBirth,
				lastContact: timeToDateStr(row.FromLastContact),
			}
		}

		if _, ok := nodeMap[row.ToPersonID]; !ok {
			nodeMap[row.ToPersonID] = nodeData{
				name:        row.ToName,
				nickname:    row.ToNickname,
				avatar:      row.ToAvatar,
				isSelf:      row.ToIsSelf,
				dateOfBirth: row.ToDateOfBirth,
				lastContact: timeToDateStr(row.ToLastContact),
			}
		}

		// Canonical direction: source = lower ID, target = higher ID.
		src, tgt := row.FromPersonID, row.ToPersonID
		if src > tgt {
			src, tgt = tgt, src
		}

		// Determine type/reverse_type based on canonical direction.
		var typeName, reverseTypeName string
		if row.FromPersonID == src {
			// This row's direction matches canonical (from < to).
			typeName = row.TypeName
			reverseTypeName = row.ReverseName

			if reverseTypeName == "" {
				reverseTypeName = typeName
			}
		} else {
			// This row is the inverse direction; swap type names.
			typeName = row.ReverseName
			reverseTypeName = row.TypeName

			if typeName == "" {
				typeName = reverseTypeName
			}
		}

		// Key includes type name so reciprocal rows for the same relationship
		// dedup (A→B and B→A produce the same canonical typeName), but two
		// distinct relationship types between the same pair produce separate links.
		key := fmt.Sprintf("%d-%d-%s", src, tgt, typeName)
		if _, exists := linkMap[key]; exists {
			continue
		}

		linkMap[key] = GraphLink{
			Source:      src,
			Target:      tgt,
			Type:        typeName,
			ReverseType: reverseTypeName,
		}
	}

	// For global graph, include all people — even those with no relationships yet.
	if personID == 0 {
		var allPersonRows []struct {
			ID          int64      `bun:"id"`
			Name        string     `bun:"name"`
			Nickname    string     `bun:"nickname"`
			AvatarPath  string     `bun:"avatar_path"`
			IsSelf      bool       `bun:"is_self"`
			DateOfBirth *string    `bun:"date_of_birth"`
			LastContact *time.Time `bun:"last_contact_at"`
		}

		if err := s.db.NewRaw(
			`SELECT id, name, COALESCE(nickname,'') AS nickname,`+
				` COALESCE(avatar_path,'') AS avatar_path, is_self, date_of_birth, last_contact_at`+
				` FROM person`,
		).Scan(ctx, &allPersonRows); err == nil {
			for _, p := range allPersonRows {
				if _, ok := nodeMap[p.ID]; !ok {
					nodeMap[p.ID] = nodeData{
						name:        p.Name,
						nickname:    p.Nickname,
						avatar:      p.AvatarPath,
						isSelf:      p.IsSelf,
						dateOfBirth: p.DateOfBirth,
						lastContact: timeToDateStr(p.LastContact),
					}
				}
			}
		}
	}

	// Batch-fetch labels for all nodes to assign groups.
	allIDs := make([]int64, 0, len(nodeMap))
	for id := range nodeMap {
		allIDs = append(allIDs, id)
	}

	var labelsByPerson map[int64][]people.Label

	if len(allIDs) > 0 {
		labelsByPerson, err = s.PeopleLabels.ListByPersonIDs(ctx, allIDs)
		if err != nil {
			return Graph{}, fmt.Errorf("relationships: graph fetch labels: %w", err)
		}
	} else {
		labelsByPerson = make(map[int64][]people.Label)
	}

	nodes := make([]GraphNode, 0, len(nodeMap))

	for id, nd := range nodeMap {
		var (
			group  string
			groups []string
		)

		if labels, ok := labelsByPerson[id]; ok && len(labels) > 0 {
			group = labels[0].Name
			groups = make([]string, len(labels))

			for i, l := range labels {
				groups[i] = l.Name
			}
		}

		nodes = append(nodes, GraphNode{
			ID:          id,
			Name:        nd.name,
			Nickname:    nd.nickname,
			Avatar:      avatarURL(id, nd.avatar),
			Group:       group,
			Groups:      groups,
			IsSelf:      nd.isSelf,
			DateOfBirth: nd.dateOfBirth,
			LastContact: nd.lastContact,
		})
	}

	links := make([]GraphLink, 0, len(linkMap))
	for _, l := range linkMap {
		links = append(links, l)
	}

	return Graph{Nodes: nodes, Links: links}, nil
}
