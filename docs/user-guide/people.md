# People

People are the core of kith-pms. Each person record holds contact info, relationships, important dates, work history, and a linked journal.

## Add a person

1. Go to **People** → **New person**
2. Fill in name (required), date of birth, relationship type
3. Save — the person appears in your list

## Edit a person

Open the person's detail page and click **Edit** in the Overview section. All fields are editable inline.

## Contacts & locations

In the detail page, use the **Contacts** and **Locations** sections to add email addresses, phone numbers, and physical addresses. Each row has inline add/edit/delete — no modal dialogs.

## Labels

Assign labels to categorize people (e.g. "family", "work", "close friend"). Labels are managed under **Settings → Labels**.

## Relationships

Link two people together with a typed relationship (e.g. "spouse", "colleague"). Relationship types are managed under **Settings → Relationship Types**.

## Avatar

Click the avatar area in edit mode to upload a profile photo. Photos are stored in `data/avatars/` (configurable via `AVATAR_STORAGE_PATH`).

## Search & filter

Use the search bar on the People list to filter by name. Use label filter pills to narrow by category.
