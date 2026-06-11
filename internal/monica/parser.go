package monica

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"
)

// stringOrFloat unmarshals a JSON value that Monica may export as either a
// bare number (2.5) or a quoted string ("2.5").
type stringOrFloat float64

func (f *stringOrFloat) UnmarshalJSON(b []byte) error {
	var n float64
	if err := json.Unmarshal(b, &n); err == nil {
		*f = stringOrFloat(n)
		return nil
	}

	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	if s == "" {
		*f = 0
		return nil
	}

	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}

	*f = stringOrFloat(n)

	return nil
}

// Export is the top-level decoded result from a Monica v4 export.
type Export struct {
	Contacts              []Contact         `json:"contacts"`
	AccountJournalEntries []MAccountJournal `json:"-"`
	// AccountDocumentCount is the number of account-level documents with no contact link (all skipped).
	AccountDocumentCount int `json:"-"`
}

type ImportOptions struct {
	ImportInactiveReminders     bool
	ImportAccountJournalEntries bool
}

// ---- v4 array-of-groups wire types ------------------------------------------

type v4Root struct {
	Account *v4Account `json:"account"`
}

type v4Account struct {
	Data       v4Groups     `json:"data"`
	Properties v4Properties `json:"properties"`
}

// v4Group is one named group in the array-of-groups wire format.
type v4Group struct {
	Type   string            `json:"type"`
	Count  int               `json:"count"`
	Values []json.RawMessage `json:"values"`
}

// v4Groups is the array-of-groups container used for both account.data and contact.data.
type v4Groups []v4Group

// values returns raw JSON values for the first group matching t, or nil if absent.
func (g v4Groups) values(t string) []json.RawMessage {
	for _, grp := range g {
		if grp.Type == t {
			return grp.Values
		}
	}

	return nil
}

// decodeGroup decodes all values in the named group into typed structs, skipping malformed entries.
func decodeGroup[T any](groups v4Groups, typeName string) []T {
	raws := groups.values(typeName)

	out := make([]T, 0, len(raws))
	for _, raw := range raws {
		var v T
		if err := json.Unmarshal(raw, &v); err != nil {
			slog.Debug("monica-import: skip malformed entry", "type", typeName, "err", err)
			continue
		}

		out = append(out, v)
	}

	return out
}

type v4Photo struct {
	UUID       string `json:"uuid"`
	Properties struct {
		DataURL string `json:"dataUrl"`
	} `json:"properties"`
}

type v4Activity struct {
	UUID       string `json:"uuid"`
	CreatedAt  string `json:"created_at"`
	Properties struct {
		Summary     string `json:"summary"`
		Description string `json:"description"`
		HappenedAt  string `json:"happened_at"`
	} `json:"properties"`
}

type v4Properties struct {
	JournalEntries []v4JournalEntry `json:"journal_entries"`
}

type v4Contact struct {
	UUID       string         `json:"uuid"`
	CreatedAt  string         `json:"created_at"`
	UpdatedAt  string         `json:"updated_at"`
	Properties v4ContactProps `json:"properties"`
	Data       v4Groups       `json:"data"`
}

type v4ContactProps struct {
	FirstName    string      `json:"first_name"`
	MiddleName   string      `json:"middle_name"`
	LastName     string      `json:"last_name"`
	Nickname     string      `json:"nickname"`
	Description  string      `json:"description"`
	Job          string      `json:"job"`
	Company      string      `json:"company"`
	IsStarred    bool        `json:"is_starred"`
	IsDead       bool        `json:"is_dead"`
	Birthdate    *v4SpecDate `json:"birthdate"`
	DeceasedDate *v4SpecDate `json:"deceased_date"`
	FirstMetDate *v4SpecDate `json:"first_met_date"`
	Tags         []string    `json:"tags"`
	// Avatar fields in real Monica 4.x exports are nested under "avatar".
	Avatar struct {
		AvatarSource string `json:"avatar_source"` // gravatar|adorable|default|external|photo
		AvatarPhoto  string `json:"avatar_photo"`  // UUID ref to account-level Photo object
	} `json:"avatar"`
}

type v4SpecDate struct {
	IsAgeBase     bool   `json:"is_age_based"`
	IsYearUnknown bool   `json:"is_year_unknown"`
	Date          string `json:"date"` // RFC3339 or "YYYY-MM-DD"
}

type v4Conversation struct {
	UUID       string `json:"uuid"`
	Properties struct {
		HappenedAt       string      `json:"happened_at"`
		ContactFieldType string      `json:"contact_field_type"`
		Messages         []v4Message `json:"messages"`
	} `json:"properties"`
}

type v4Message struct {
	Properties struct {
		Content     string `json:"content"`
		WrittenAt   string `json:"written_at"`
		WrittenByMe bool   `json:"written_by_me"`
	} `json:"properties"`
}

type v4LifeEvent struct {
	UUID       string `json:"uuid"`
	Properties struct {
		Name       string `json:"name"`
		Note       string `json:"note"`
		HappenedAt string `json:"happened_at"`
		Type       string `json:"type"`
	} `json:"properties"`
}

type v4Note struct {
	UUID       string `json:"uuid"`
	CreatedAt  string `json:"created_at"`
	Properties struct {
		Body string `json:"body"`
	} `json:"properties"`
}

type v4Reminder struct {
	UUID       string `json:"uuid"`
	Properties struct {
		InitialDate   string `json:"initial_date"`
		Title         string `json:"title"`
		Description   string `json:"description"`
		FrequencyType string `json:"frequency_type"`
		Inactive      bool   `json:"inactive"`
	} `json:"properties"`
}

type v4Address struct {
	UUID       string `json:"uuid"`
	Properties struct {
		Name       string `json:"name"`
		Street     string `json:"street"`
		City       string `json:"city"`
		Province   string `json:"province"`
		PostalCode string `json:"postal_code"`
		Country    string `json:"country"`
	} `json:"properties"`
}

type v4ContactField struct {
	UUID       string `json:"uuid"`
	Properties struct {
		Data string `json:"data"`
		Type string `json:"type"` // UUID of ContactFieldType
	} `json:"properties"`
}

type v4Call struct {
	UUID       string `json:"uuid"`
	CreatedAt  string `json:"called_at"`
	Properties struct {
		CalledAt string `json:"called_at"`
		Content  string `json:"content"`
	} `json:"properties"`
}

type v4Task struct {
	UUID       string `json:"uuid"`
	Properties struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Completed   bool   `json:"completed"`
		CompletedAt string `json:"completed_at"`
	} `json:"properties"`
}

type v4Gift struct {
	UUID       string `json:"uuid"`
	Properties struct {
		Name    string        `json:"name"`
		Comment string        `json:"comment"`
		URL     string        `json:"url"`
		Amount  stringOrFloat `json:"amount"`
		Status  string        `json:"status"`
		Date    string        `json:"date"`
	} `json:"properties"`
}

type v4Relationship struct {
	UUID       string `json:"uuid"`
	Properties struct {
		Type      string `json:"type"`       // relationship type name
		ContactIs string `json:"contact_is"` // UUID of from-contact
		OfContact string `json:"of_contact"` // UUID of to-contact
	} `json:"properties"`
}

type v4Document struct {
	UUID       string `json:"uuid"`
	Properties struct {
		OriginalFilename string `json:"original_filename"`
		Filesize         int64  `json:"filesize"`
		Type             string `json:"type"`
		MimeType         string `json:"mime_type"`
		DataURL          string `json:"dataUrl"`
	} `json:"properties"`
}

// UnmarshalJSON tolerates Monica exports that store document entries as bare
// strings (UUID references) rather than full objects; those are left empty
// and filtered out downstream by the DataURL check.
func (d *v4Document) UnmarshalJSON(b []byte) error {
	if len(b) > 0 && b[0] == '"' {
		return nil
	}

	type plain v4Document

	return json.Unmarshal(b, (*plain)(d))
}

type v4JournalEntry struct {
	UUID       string `json:"uuid"`
	CreatedAt  string `json:"created_at"`
	Properties struct {
		Title string `json:"title"`
		Post  string `json:"post"`
		Date  string `json:"date"`
	} `json:"properties"`
}

// ---- normalized domain types ------------------------------------------------

// Contact is the normalised contact used by the mapper.
type Contact struct {
	ID          string         `json:"id"`
	FirstName   string         `json:"first_name"`
	MiddleName  string         `json:"middle_name"`
	LastName    string         `json:"last_name"`
	Nickname    string         `json:"nickname"`
	Description string         `json:"description"`
	Company     string         `json:"company"`
	Job         string         `json:"job"`
	Information Information    `json:"information"`
	Addresses   []Address      `json:"addresses"`
	ContactInfo []ContactField `json:"contactInformation"`
	Notes       []Note         `json:"notes"`
	Activities  []MActivity    `json:"activities"`
	Reminders   []MReminder    `json:"reminders"`
	Tags        []Tag          `json:"tags"`
	Calls       []MCall        `json:"calls"`
	Tasks       []MTask        `json:"tasks"`
	Gifts       []MGift        `json:"gifts"`
	Documents   []MDocument    `json:"-"`
	// v4 relationship data resolved at parse time
	Relationships []MRelationship `json:"-"`
	// AvatarDataURL is a "data:<mime>;base64,..." string resolved from account photos.
	// Only set when avatar_source == "photo" and the referenced photo UUID is found.
	AvatarDataURL string          `json:"-"`
	Conversations []MConversation `json:"-"`
	LifeEvents    []MLifeEvent    `json:"-"`
}

type Information struct {
	Birthdate     string `json:"birthdate"` // "YYYY-MM-DD" | "0000-MM-DD" | ""
	IsYearUnknown bool   `json:"is_year_unknown"`
	FirstMetDate  string `json:"first_met_date"` // "YYYY-MM-DD" | ""
}

type Address struct {
	Name       string `json:"name"`
	Street     string `json:"street"`
	City       string `json:"city"`
	Province   string `json:"province"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

type ContactField struct {
	Data string           `json:"data"`
	Type ContactFieldType `json:"contact_field_type"`
}

type ContactFieldType struct {
	Name string `json:"name"`
}

type Note struct {
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"` // ISO 8601
}

type MActivity struct {
	UUID        string `json:"uuid"`
	Summary     string `json:"summary"`
	Description string `json:"description"`
	HappenedAt  string `json:"happened_at"` // "YYYY-MM-DD"
}

type MReminder struct {
	Title         string `json:"title"`
	Description   string `json:"description"`
	InitialDate   string `json:"initial_date"` // "YYYY-MM-DD"
	FrequencyType string `json:"frequency_type"`
	Inactive      bool   `json:"inactive"`
}

type Tag struct {
	Name string `json:"name"`
}

type MCall struct {
	CalledAt string `json:"called_at"` // ISO 8601
	Content  string `json:"content"`
}

type MTask struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
	CompletedAt string `json:"completed_at"`
}

type MGift struct {
	Name    string  `json:"name"`
	Comment string  `json:"comment"`
	Amount  float64 `json:"amount"`
	Status  string  `json:"status"`
	Date    string  `json:"date"`
}

type MDocument struct {
	OriginalFilename string
	MimeType         string
	Filesize         int64
	DataURL          string
}

type MRelationship struct {
	TypeName      string `json:"type_name"`
	ToContactUUID string `json:"to_contact_uuid"`
	ToContactName string `json:"to_contact_name"` // resolved after all contacts parsed
}

type MConversation struct {
	HappenedAt string
	Channel    string // contact_field_type UUID; blank = unresolved
	Messages   []MMessage
}

type MMessage struct {
	Content     string
	WrittenAt   string // ISO 8601
	WrittenByMe bool
}

type MLifeEvent struct {
	Name       string
	Note       string
	HappenedAt string
	Type       string // life_event_type UUID
}

type MAccountJournal struct {
	Title          string `json:"title"`
	Content        string `json:"content"`
	OccurredAtDate string `json:"occurred_at_date"`
}

// Parse decodes a Monica v4 JSON export from r.
func Parse(r io.Reader) (*Export, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var root v4Root
	if err := json.Unmarshal(raw, &root); err != nil {
		return nil, err
	}

	if root.Account == nil {
		return nil, fmt.Errorf(
			"monica-import: unrecognised export format (expected Monica v4 with top-level \"account\" key)",
		)
	}

	return parseV4(root.Account)
}

func parseV4(acc *v4Account) (*Export, error) {
	contacts := decodeGroup[v4Contact](acc.Data, "contact")

	// Build UUID→name lookup for relationship resolution.
	uuidToName := make(map[string]string, len(contacts))
	for _, c := range contacts {
		name := c.Properties.FirstName
		if c.Properties.LastName != "" {
			if name != "" {
				name += " "
			}

			name += c.Properties.LastName
		}

		if name == "" {
			name = c.Properties.Nickname
		}

		uuidToName[c.UUID] = name
	}

	// Build UUID→[]MRelationship from account-level relationships.
	rawRels := decodeGroup[v4Relationship](acc.Data, "relationship")

	uuidToRels := make(map[string][]MRelationship, len(rawRels))
	for _, r := range rawRels {
		p := r.Properties
		if p.ContactIs == "" || p.OfContact == "" {
			continue
		}

		uuidToRels[p.ContactIs] = append(uuidToRels[p.ContactIs], MRelationship{
			TypeName:      p.Type,
			ToContactUUID: p.OfContact,
			ToContactName: uuidToName[p.OfContact],
		})
	}

	// Build UUID→dataURL map from account-level photos for avatar resolution.
	rawPhotos := decodeGroup[v4Photo](acc.Data, "photo")

	photoURLs := make(map[string]string, len(rawPhotos))
	for _, p := range rawPhotos {
		if p.UUID != "" && p.Properties.DataURL != "" {
			photoURLs[p.UUID] = p.Properties.DataURL
		}
	}

	// Build UUID→MActivity map from account-level activities for per-contact resolution.
	rawActivities := decodeGroup[v4Activity](acc.Data, "activity")

	activityByUUID := make(map[string]MActivity, len(rawActivities))
	for _, a := range rawActivities {
		if a.UUID == "" {
			continue
		}

		ap := a.Properties

		title := strings.TrimSpace(ap.Summary)
		if title == "" {
			title = truncate(ap.Description, 60)
		}

		if title == "" {
			continue
		}

		activityByUUID[a.UUID] = MActivity{
			Summary:     title,
			Description: ap.Description,
			HappenedAt:  normDate(ap.HappenedAt),
		}
	}

	// Count account-level documents (no contact link; all are skipped by design).
	accountDocCount := len(acc.Data.values("document"))

	normalContacts := make([]Contact, 0, len(contacts))
	for _, c := range contacts {
		normalContacts = append(normalContacts, normaliseV4Contact(c, uuidToRels[c.UUID], photoURLs, activityByUUID))
	}

	return &Export{
		Contacts:              normalContacts,
		AccountJournalEntries: normaliseV4AccountJournal(acc.Properties.JournalEntries),
		AccountDocumentCount:  accountDocCount,
	}, nil
}

func normaliseV4AccountJournal(entries []v4JournalEntry) []MAccountJournal {
	out := make([]MAccountJournal, 0, len(entries))
	for _, entry := range entries {
		p := entry.Properties

		title := p.Title
		if title == "" {
			title = "Journal entry"
		}

		content := p.Post
		if content == "" {
			continue
		}

		occurredAt := p.Date
		if occurredAt == "" {
			occurredAt = dateFromISO(entry.CreatedAt)
		}

		out = append(out, MAccountJournal{
			Title:          title,
			Content:        content,
			OccurredAtDate: occurredAt,
		})
	}

	return out
}

func normaliseV4Contact(
	c v4Contact,
	rels []MRelationship,
	photoURLs map[string]string,
	activityByUUID map[string]MActivity,
) Contact {
	p := c.Properties

	info := Information{}
	if p.Birthdate != nil && p.Birthdate.Date != "" {
		info.Birthdate = normDate(p.Birthdate.Date)
		info.IsYearUnknown = p.Birthdate.IsYearUnknown
	}

	if p.FirstMetDate != nil && p.FirstMetDate.Date != "" {
		info.FirstMetDate = normDate(p.FirstMetDate.Date)
	}

	tags := make([]Tag, 0, len(p.Tags))
	for _, t := range p.Tags {
		tags = append(tags, Tag{Name: t})
	}

	rawAddresses := decodeGroup[v4Address](c.Data, "address")

	addresses := make([]Address, 0, len(rawAddresses))
	for _, a := range rawAddresses {
		ap := a.Properties
		addresses = append(addresses, Address{
			Name:       ap.Name,
			Street:     ap.Street,
			City:       ap.City,
			Province:   ap.Province,
			PostalCode: ap.PostalCode,
			Country:    ap.Country,
		})
	}

	rawCFs := decodeGroup[v4ContactField](c.Data, "contact_field")

	contactInfo := make([]ContactField, 0, len(rawCFs))
	for _, cf := range rawCFs {
		if cf.Properties.Data == "" {
			continue
		}
		// v4 stores type as UUID; we use the UUID as the name — mapper will treat unknown as "other"
		contactInfo = append(contactInfo, ContactField{
			Data: cf.Properties.Data,
			Type: ContactFieldType{Name: cf.Properties.Type},
		})
	}

	rawNotes := decodeGroup[v4Note](c.Data, "note")

	notes := make([]Note, 0, len(rawNotes))
	for _, n := range rawNotes {
		if n.Properties.Body == "" {
			continue
		}

		notes = append(notes, Note{
			Body:      n.Properties.Body,
			CreatedAt: n.CreatedAt,
		})
	}

	rawReminders := decodeGroup[v4Reminder](c.Data, "reminder")

	reminders := make([]MReminder, 0, len(rawReminders))
	for _, r := range rawReminders {
		rp := r.Properties
		if rp.Title == "" || rp.InitialDate == "" {
			continue
		}

		reminders = append(reminders, MReminder{
			Title:         rp.Title,
			Description:   rp.Description,
			InitialDate:   normDate(rp.InitialDate),
			FrequencyType: rp.FrequencyType,
			Inactive:      rp.Inactive,
		})
	}

	rawCalls := decodeGroup[v4Call](c.Data, "call")

	calls := make([]MCall, 0, len(rawCalls))
	for _, call := range rawCalls {
		cp := call.Properties

		ts := cp.CalledAt
		if ts == "" {
			ts = call.CreatedAt
		}

		calls = append(calls, MCall{
			CalledAt: ts,
			Content:  cp.Content,
		})
	}

	rawTasks := decodeGroup[v4Task](c.Data, "task")

	tasks := make([]MTask, 0, len(rawTasks))
	for _, task := range rawTasks {
		tp := task.Properties
		if tp.Title == "" {
			continue
		}

		tasks = append(tasks, MTask{
			Title:       tp.Title,
			Description: tp.Description,
			Completed:   tp.Completed,
			CompletedAt: tp.CompletedAt,
		})
	}

	rawGifts := decodeGroup[v4Gift](c.Data, "gift")

	gifts := make([]MGift, 0, len(rawGifts))
	for _, g := range rawGifts {
		gp := g.Properties
		if gp.Name == "" {
			continue
		}

		gifts = append(gifts, MGift{
			Name:    gp.Name,
			Comment: gp.Comment,
			Amount:  float64(gp.Amount),
			Status:  gp.Status,
			Date:    normDate(gp.Date),
		})
	}

	rawDocs := decodeGroup[v4Document](c.Data, "document")

	documents := make([]MDocument, 0, len(rawDocs))
	for _, d := range rawDocs {
		dp := d.Properties
		if dp.DataURL == "" {
			continue
		}

		documents = append(documents, MDocument{
			OriginalFilename: dp.OriginalFilename,
			MimeType:         dp.MimeType,
			Filesize:         dp.Filesize,
			DataURL:          dp.DataURL,
		})
	}

	// Resolve avatar: only import when source is "photo" and UUID resolves to a dataUrl.
	var avatarDataURL string
	if p.Avatar.AvatarSource == "photo" && p.Avatar.AvatarPhoto != "" {
		avatarDataURL = photoURLs[p.Avatar.AvatarPhoto]
		if avatarDataURL == "" {
			slog.Debug(
				"monica-import: avatar photo UUID not found in export",
				"contact", c.UUID,
				"photo_uuid", p.Avatar.AvatarPhoto,
			)
		}
	}

	rawConvs := decodeGroup[v4Conversation](c.Data, "conversation")

	conversations := make([]MConversation, 0, len(rawConvs))
	for _, conv := range rawConvs {
		cp := conv.Properties

		msgs := make([]MMessage, 0, len(cp.Messages))
		for _, m := range cp.Messages {
			msgs = append(msgs, MMessage{
				Content:     m.Properties.Content,
				WrittenAt:   m.Properties.WrittenAt,
				WrittenByMe: m.Properties.WrittenByMe,
			})
		}

		if len(msgs) == 0 {
			continue
		}

		conversations = append(conversations, MConversation{
			HappenedAt: cp.HappenedAt,
			Channel:    cp.ContactFieldType,
			Messages:   msgs,
		})
	}

	rawLEs := decodeGroup[v4LifeEvent](c.Data, "life_event")

	lifeEvents := make([]MLifeEvent, 0, len(rawLEs))
	for _, le := range rawLEs {
		lp := le.Properties
		if strings.TrimSpace(lp.Name) == "" {
			continue
		}

		lifeEvents = append(lifeEvents, MLifeEvent{
			Name:       lp.Name,
			Note:       lp.Note,
			HappenedAt: lp.HappenedAt,
			Type:       lp.Type,
		})
	}

	// Resolve per-contact activity UUIDs against the account-level activity map.
	rawActivityUUIDs := c.Data.values("activity")

	activities := make([]MActivity, 0, len(rawActivityUUIDs))
	for _, raw := range rawActivityUUIDs {
		var uuid string
		if err := json.Unmarshal(raw, &uuid); err != nil || uuid == "" {
			continue
		}

		if act, ok := activityByUUID[uuid]; ok {
			act.UUID = uuid
			activities = append(activities, act)
		}
	}

	return Contact{
		ID:            c.UUID,
		FirstName:     p.FirstName,
		MiddleName:    p.MiddleName,
		LastName:      p.LastName,
		Nickname:      p.Nickname,
		Description:   p.Description,
		Company:       p.Company,
		Job:           p.Job,
		Information:   info,
		Addresses:     addresses,
		ContactInfo:   contactInfo,
		Notes:         notes,
		Activities:    activities,
		Reminders:     reminders,
		Tags:          tags,
		Calls:         calls,
		Tasks:         tasks,
		Gifts:         gifts,
		Documents:     documents,
		Relationships: rels,
		AvatarDataURL: avatarDataURL,
		Conversations: conversations,
		LifeEvents:    lifeEvents,
	}
}

// normDate normalizes an RFC3339 timestamp to "YYYY-MM-DD" by trimming.
// Strings shorter than 10 chars (empty, partial) are returned unchanged.
func normDate(s string) string {
	if len(s) > 10 {
		return s[:10]
	}

	return s
}
