package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/auth"
	"github.com/nhymxu/kith-pms/internal/people"
	"github.com/nhymxu/kith-pms/internal/reminders"
	"github.com/nhymxu/kith-pms/internal/web/templates"
)

type RemindersHandlers struct {
	Svc       *reminders.Service
	PeopleSvc *people.Service
}

func (h *RemindersHandlers) GetList(c *echo.Context) error {
	status := c.QueryParam("status")
	if status == "" {
		status = "pending"
	}

	var personID *int64

	if pidStr := c.QueryParam("person_id"); pidStr != "" {
		if pid, err := strconv.ParseInt(pidStr, 10, 64); err == nil {
			personID = &pid
		}
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	list, err := h.Svc.List(c.Request().Context(), reminders.ListParams{
		Status:   status,
		PersonID: personID,
		PageSize: 50,
		Page:     page,
	})
	if err != nil {
		return err
	}

	hasMore := len(list) == 50

	component := templates.RemindersList(templates.RemindersListParams{
		Reminders: list,
		Status:    status,
		Page:      page,
		HasMore:   hasMore,
		CSRFToken: auth.CSRFToken(c),
	})

	return component.Render(c.Request().Context(), c.Response())
}

func (h *RemindersHandlers) GetNew(c *echo.Context) error {
	var allPeople []people.Person
	if h.PeopleSvc != nil {
		allPeople, _ = h.PeopleSvc.List(c.Request().Context(), people.ListParams{PageSize: 1000, Page: 1})
	}

	component := templates.RemindersForm(templates.RemindersFormParams{
		CSRFToken: auth.CSRFToken(c),
		AllPeople: allPeople,
	})

	return component.Render(c.Request().Context(), c.Response())
}

func (h *RemindersHandlers) PostCreate(c *echo.Context) error {
	rem, formErr := parseReminderForm(c)
	if formErr != "" {
		var allPeople []people.Person
		if h.PeopleSvc != nil {
			allPeople, _ = h.PeopleSvc.List(c.Request().Context(), people.ListParams{PageSize: 1000, Page: 1})
		}

		component := templates.RemindersForm(templates.RemindersFormParams{
			Reminder:  rem,
			CSRFToken: auth.CSRFToken(c),
			Error:     formErr,
			AllPeople: allPeople,
		})

		return component.Render(c.Request().Context(), c.Response())
	}

	id, err := h.Svc.Create(c.Request().Context(), &rem)
	if err != nil {
		return err
	}

	return c.Redirect(http.StatusSeeOther, "/reminders/"+strconv.FormatInt(id, 10))
}

func (h *RemindersHandlers) GetDetail(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return echo.ErrNotFound
	}

	rem, err := h.Svc.GetByID(c.Request().Context(), id)
	if err != nil {
		return err
	}

	var personName string

	if rem.PersonID != nil && h.PeopleSvc != nil {
		if p, err := h.PeopleSvc.Get(c.Request().Context(), *rem.PersonID); err == nil && p != nil {
			personName = p.Name
		}
	}

	component := templates.RemindersDetail(templates.RemindersDetailParams{
		Reminder:   *rem,
		PersonName: personName,
		CSRFToken:  auth.CSRFToken(c),
	})

	return component.Render(c.Request().Context(), c.Response())
}

func (h *RemindersHandlers) GetEdit(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return echo.ErrNotFound
	}

	rem, err := h.Svc.GetByID(c.Request().Context(), id)
	if err != nil {
		return err
	}

	var allPeople []people.Person
	if h.PeopleSvc != nil {
		allPeople, _ = h.PeopleSvc.List(c.Request().Context(), people.ListParams{PageSize: 1000, Page: 1})
	}

	component := templates.RemindersForm(templates.RemindersFormParams{
		Reminder:  *rem,
		CSRFToken: auth.CSRFToken(c),
		IsEdit:    true,
		AllPeople: allPeople,
	})

	return component.Render(c.Request().Context(), c.Response())
}

func (h *RemindersHandlers) PutUpdate(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return echo.ErrNotFound
	}

	rem, formErr := parseReminderForm(c)
	rem.ID = id

	if formErr != "" {
		var allPeople []people.Person
		if h.PeopleSvc != nil {
			allPeople, _ = h.PeopleSvc.List(c.Request().Context(), people.ListParams{PageSize: 1000, Page: 1})
		}

		component := templates.RemindersForm(templates.RemindersFormParams{
			Reminder:  rem,
			CSRFToken: auth.CSRFToken(c),
			IsEdit:    true,
			Error:     formErr,
			AllPeople: allPeople,
		})

		return component.Render(c.Request().Context(), c.Response())
	}

	if err := h.Svc.Update(c.Request().Context(), &rem); err != nil {
		return err
	}

	return c.Redirect(http.StatusSeeOther, "/reminders/"+strconv.FormatInt(id, 10))
}

func (h *RemindersHandlers) Delete(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return echo.ErrNotFound
	}

	if err := h.Svc.Delete(c.Request().Context(), id); err != nil {
		return err
	}

	return c.Redirect(http.StatusSeeOther, "/reminders")
}

func (h *RemindersHandlers) PostToggleComplete(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return echo.ErrNotFound
	}

	if err := h.Svc.MarkComplete(c.Request().Context(), id); err != nil {
		return err
	}

	rem, err := h.Svc.GetByID(c.Request().Context(), id)
	if err != nil {
		return err
	}

	var personName string

	if rem.PersonID != nil && h.PeopleSvc != nil {
		if p, err := h.PeopleSvc.Get(c.Request().Context(), *rem.PersonID); err == nil && p != nil {
			personName = p.Name
		}
	}

	rwp := reminders.ReminderWithPerson{
		Reminder:   *rem,
		PersonName: personName,
	}

	component := templates.ReminderRow(rwp, auth.CSRFToken(c))

	return component.Render(c.Request().Context(), c.Response())
}

func parseReminderForm(c *echo.Context) (reminders.Reminder, string) {
	var rem reminders.Reminder

	rem.Title = c.FormValue("title")
	rem.Notes = c.FormValue("notes")

	dueDateStr := c.FormValue("due_date")
	if dueDateStr == "" {
		return rem, "Due date is required"
	}

	dueDate, err := time.Parse("2006-01-02", dueDateStr)
	if err != nil {
		return rem, "Invalid due date format"
	}

	rem.DueDate = dueDate

	if personIDStr := c.FormValue("person_id"); personIDStr != "" {
		if pid, err := strconv.ParseInt(personIDStr, 10, 64); err == nil && pid > 0 {
			rem.PersonID = &pid
		}
	}

	if c.FormValue("completed") == "on" {
		rem.Completed = true
		now := time.Now()
		rem.CompletedAt = &now
	}

	if rem.Title == "" {
		return rem, "Title is required"
	}

	return rem, ""
}
