package monica

import (
	"encoding/json"
	"fmt"
	"io"
)

// Export is the top-level decoded result from a Monica v4 export:
//
//	{"account": {"data": {"contacts": [...]}}}
type Export struct {
	Contacts              []Contact         `json:"contacts"`
	AccountJournalEntries []MAccountJournal `json:"-"`
}

type ImportOptions struct {
	ImportInactiveReminders     bool
	ImportAccountJournalEntries bool
}

// ---- v4 wire types ----------------------------------------------------------

type v4Root struct {
	Account *v4Account `json:"account"`
}

type v4Account struct {
	Data       v4Data       `json:"data"`
	Properties v4Properties `json:"properties"`
}

type v4Data struct {
	Contacts      []v4Contact      `json:"contacts"`
	Relationships []v4Relationship `json:"relationships"`
}

type v4Properties struct {
	JournalEntries []v4JournalEntry `json:"journal_entries"`
}

type v4Contact struct {
	UUID       string         `json:"uuid"`
	CreatedAt  string         `json:"created_at"`
	UpdatedAt  string         `json:"updated_at"`
	Properties v4ContactProps `json:"properties"`
	Data       v4ContactData  `json:"data"`
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
}

type v4SpecDate struct {
	IsAgeBase     bool   `json:"is_age_based"`
	IsYearUnknown bool   `json:"is_year_unknown"`
	Date          string `json:"date"` // "YYYY-MM-DD"
}

type v4ContactData struct {
	Notes         v4CountColl[v4Note]         `json:"notes"`
	Activities    v4UUIDColl                  `json:"activities"`
	Reminders     v4CountColl[v4Reminder]     `json:"reminders"`
	Addresses     v4CountColl[v4Address]      `json:"addresses"`
	ContactFields v4CountColl[v4ContactField] `json:"contact_fields"`
	Calls         v4CountColl[v4Call]         `json:"calls"`
	Tasks         v4CountColl[v4Task]         `json:"tasks"`
	Gifts         v4CountColl[v4Gift]         `json:"gifts"`
}

// v4CountColl is a collection that embeds items directly.
type v4CountColl[T any] struct {
	Data []T `json:"data"`
}

// v4UUIDColl is a collection of UUID strings only.
type v4UUIDColl struct {
	Data []string `json:"data"`
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
		Name    string  `json:"name"`
		Comment string  `json:"comment"`
		URL     string  `json:"url"`
		Amount  float64 `json:"amount"`
		Status  string  `json:"status"`
		Date    string  `json:"date"`
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

type v4JournalEntry struct {
	UUID       string `json:"uuid"`
	CreatedAt  string `json:"created_at"`
	Properties struct {
		Title string `json:"title"`
		Post  string `json:"post"`
		Date  string `json:"date"`
	} `json:"properties"`
}

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
	// v4 relationship data resolved at parse time
	Relationships []MRelationship `json:"-"`
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

type MRelationship struct {
	TypeName      string `json:"type_name"`
	ToContactUUID string `json:"to_contact_uuid"`
	ToContactName string `json:"to_contact_name"` // resolved after all contacts parsed
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
		return nil, fmt.Errorf("monica-import: unrecognised export format (expected Monica v4 with top-level \"account\" key)")
	}

	return parseV4(root.Account)
}

func parseV4(acc *v4Account) (*Export, error) {
	// Build UUID→name lookup for relationship resolution
	uuidToName := make(map[string]string, len(acc.Data.Contacts))
	for _, c := range acc.Data.Contacts {
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

	// Build UUID→[]MRelationship from account-level relationships
	uuidToRels := make(map[string][]MRelationship, len(acc.Data.Relationships))
	for _, r := range acc.Data.Relationships {
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

	contacts := make([]Contact, 0, len(acc.Data.Contacts))
	for _, c := range acc.Data.Contacts {
		contacts = append(contacts, normaliseV4Contact(c, uuidToRels[c.UUID]))
	}

	return &Export{
		Contacts:              contacts,
		AccountJournalEntries: normaliseV4AccountJournal(acc.Properties.JournalEntries),
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

func normaliseV4Contact(c v4Contact, rels []MRelationship) Contact {
	p := c.Properties

	info := Information{}
	if p.Birthdate != nil && p.Birthdate.Date != "" {
		info.Birthdate = p.Birthdate.Date
		info.IsYearUnknown = p.Birthdate.IsYearUnknown
	}

	if p.FirstMetDate != nil && p.FirstMetDate.Date != "" {
		info.FirstMetDate = p.FirstMetDate.Date
	}

	tags := make([]Tag, 0, len(p.Tags))
	for _, t := range p.Tags {
		tags = append(tags, Tag{Name: t})
	}

	addresses := make([]Address, 0, len(c.Data.Addresses.Data))
	for _, a := range c.Data.Addresses.Data {
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

	contactInfo := make([]ContactField, 0, len(c.Data.ContactFields.Data))
	for _, cf := range c.Data.ContactFields.Data {
		if cf.Properties.Data == "" {
			continue
		}
		// v4 stores type as UUID; we use the UUID as the name — mapper will treat unknown as "other"
		contactInfo = append(contactInfo, ContactField{
			Data: cf.Properties.Data,
			Type: ContactFieldType{Name: cf.Properties.Type},
		})
	}

	notes := make([]Note, 0, len(c.Data.Notes.Data))
	for _, n := range c.Data.Notes.Data {
		if n.Properties.Body == "" {
			continue
		}

		notes = append(notes, Note{
			Body:      n.Properties.Body,
			CreatedAt: n.CreatedAt,
		})
	}

	reminders := make([]MReminder, 0, len(c.Data.Reminders.Data))
	for _, r := range c.Data.Reminders.Data {
		rp := r.Properties
		if rp.Title == "" || rp.InitialDate == "" {
			continue
		}

		reminders = append(reminders, MReminder{
			Title:         rp.Title,
			Description:   rp.Description,
			InitialDate:   rp.InitialDate,
			FrequencyType: rp.FrequencyType,
			Inactive:      rp.Inactive,
		})
	}

	calls := make([]MCall, 0, len(c.Data.Calls.Data))
	for _, call := range c.Data.Calls.Data {
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

	tasks := make([]MTask, 0, len(c.Data.Tasks.Data))
	for _, task := range c.Data.Tasks.Data {
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

	gifts := make([]MGift, 0, len(c.Data.Gifts.Data))
	for _, g := range c.Data.Gifts.Data {
		gp := g.Properties
		if gp.Name == "" {
			continue
		}

		gifts = append(gifts, MGift{
			Name:    gp.Name,
			Comment: gp.Comment,
			Amount:  gp.Amount,
			Status:  gp.Status,
			Date:    gp.Date,
		})
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
		Activities:    nil, // v4 activities are account-level; per-contact only has UUIDs
		Reminders:     reminders,
		Tags:          tags,
		Calls:         calls,
		Tasks:         tasks,
		Gifts:         gifts,
		Relationships: rels,
	}
}
