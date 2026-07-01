package people

import "github.com/nhymxu/kith-pms/internal/audit"

// diffPersonFields returns audit.Change entries for every profile field that
// differs between old and updated. Avatar and system fields are excluded since
// they have dedicated detail_action values.
func diffPersonFields(old, updated Person) []audit.Change {
	var ch []audit.Change

	track := func(field, oldVal, newVal string) {
		if oldVal != newVal {
			ch = append(ch, audit.Change{Field: field, Old: oldVal, New: newVal})
		}
	}

	track("name", old.Name, updated.Name)
	track("nickname", old.Nickname, updated.Nickname)
	track("prefix", old.Prefix, updated.Prefix)
	track("gender", old.Gender, updated.Gender)
	track("other_notes", old.OtherNotes, updated.OtherNotes)
	track("date_of_birth", dobStr(old.DateOfBirth), dobStr(updated.DateOfBirth))

	return ch
}

func dobStr(d *DateOnly) string {
	if d == nil {
		return ""
	}

	return d.String()
}
