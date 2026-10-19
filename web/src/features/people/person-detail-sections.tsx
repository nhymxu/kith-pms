import {
	useMutation,
	useQuery,
	useQueryClient,
	useSuspenseQuery,
} from "@tanstack/react-query";
import { Pencil, Star, X } from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { QueryBoundary } from "#/components/query-boundary";
import { Alert, AlertDescription } from "#/components/ui/alert";
import { Badge } from "#/components/ui/badge";
import { Button } from "#/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "#/components/ui/dialog";
import { Input } from "#/components/ui/input";
import { Label } from "#/components/ui/label";
import { RadioGroup, RadioGroupItem } from "#/components/ui/radio-group";
import { Switch } from "#/components/ui/switch";
import { Textarea } from "#/components/ui/textarea";
import {
	archivePerson,
	deletePerson,
	getPerson,
	restorePerson,
	setFavorite,
	unarchivePerson,
	unsetFavorite,
	updatePerson,
} from "#/endpoints/people";
import { getSettings } from "#/endpoints/settings";
import { apiErrorMessage } from "#/lib/api-error-message";
import {
	datetimeLocalToUtc,
	daysRemaining,
	formatDate,
	utcToDatetimeLocal,
} from "#/lib/format-datetime";
import { formatPersonName } from "#/lib/format-person-name";
import { keys } from "#/query-keys";
import type { Person } from "#/schemas/person";
import { genderOptions } from "#/schemas/person";
import { AvatarUploader } from "./avatar-uploader";
import { ContactsSection } from "./person-section-contacts";
import { ImportantDatesSection } from "./person-section-dates";
import { GiftsSection } from "./person-section-gifts";
import { JournalSection } from "./person-section-journal";
import { LabelsSection } from "./person-section-labels";
import { LocationsSection } from "./person-section-locations";
import { NotesSection } from "./person-section-notes";
import { PersonSectionRelationshipGraph } from "./person-section-relationship-graph";
import { RelationshipsSection } from "./person-section-relationships";
import { WorkHistorySection } from "./person-section-work-history";
import { QuickActions } from "./quick-actions";

function SectionCard({ children }: { children: React.ReactNode }) {
	return (
		<div className="rounded-base border-bw border-line bg-panel shadow-shadow p-4">
			{children}
		</div>
	);
}

function SectionHeading({ children }: { children: React.ReactNode }) {
	return (
		<h2 className="text-[11px] font-semibold uppercase tracking-widest text-sub mb-2">
			{children}
		</h2>
	);
}

interface OverviewSectionProps {
	person: Person;
	editing: boolean;
	onEdit: () => void;
	onCancel: () => void;
}

function OverviewSection({
	person,
	editing,
	onEdit,
	onCancel,
}: OverviewSectionProps) {
	const { t } = useTranslation("people");
	const qc = useQueryClient();
	const [name, setName] = useState(person.name);
	const [nickname, setNickname] = useState(person.nickname);
	const [gender, setGender] = useState(person.gender ?? "");
	const [dob, setDob] = useState(person.date_of_birth ?? "");
	const [birthdayReminder, setBirthdayReminder] = useState(
		person.has_birthday_reminder,
	);
	const [lastContactAt, setLastContactAt] = useState(
		utcToDatetimeLocal(person.last_contact_at),
	);
	const [bio, setBio] = useState(person.bio);

	const saveMutation = useMutation({
		mutationFn: () =>
			updatePerson(person.id, {
				name,
				nickname,
				gender,
				date_of_birth: dob,
				create_birthday_reminder: birthdayReminder,
				last_contact_at: lastContactAt
					? datetimeLocalToUtc(lastContactAt)
					: null,
				bio,
				contacts: person.contacts.map((c, i) => ({
					type: c.type,
					value: c.value,
					label: c.label,
					position: i,
				})),
				locations: person.locations.map((l, i) => ({
					type: l.type,
					address: l.address,
					city: l.city,
					country: l.country,
					postal_code: l.postal_code,
					position: i,
				})),
			}),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: keys.people.detail(person.id) });
			onCancel();
		},
	});

	return (
		<div className="space-y-3">
			<div className="flex items-center justify-between">
				<SectionHeading>{t("detail.overview.title")}</SectionHeading>
				{editing ? (
					<div className="flex gap-2">
						<Button variant="neutral" size="sm" onClick={onCancel}>
							{t("detail.overview.cancel")}
						</Button>
						<Button
							size="sm"
							disabled={saveMutation.isPending}
							onClick={() => saveMutation.mutate()}
						>
							{saveMutation.isPending
								? t("detail.overview.saving")
								: t("detail.overview.save")}
						</Button>
					</div>
				) : (
					<Button variant="neutral" size="sm" onClick={onEdit}>
						<Pencil className="size-3" /> {t("detail.overview.edit")}
					</Button>
				)}
			</div>

			<AvatarUploader
				personId={person.id}
				hasAvatar={Boolean(person.avatar_path)}
				showControls={editing}
			/>

			{editing ? (
				<div className="space-y-3">
					<div>
						<Label>{t("detail.overview.fullName")}</Label>
						<Input
							value={name}
							onChange={(e) => setName(e.target.value)}
							placeholder={t("detail.overview.fullName")}
						/>
					</div>
					<div>
						<Label>{t("detail.overview.nickname")}</Label>
						<Input
							value={nickname}
							onChange={(e) => setNickname(e.target.value)}
							placeholder={t("detail.overview.nickname")}
						/>
					</div>
					<div>
						<Label>{t("detail.overview.dateOfBirth")}</Label>
						<Input
							type="date"
							value={dob}
							onChange={(e) => setDob(e.target.value)}
						/>
					</div>
					{dob && (
						<div className="flex items-center gap-2">
							<Switch
								id="birthday-reminder-toggle"
								checked={birthdayReminder}
								onCheckedChange={setBirthdayReminder}
							/>
							<Label
								htmlFor="birthday-reminder-toggle"
								className="font-normal cursor-pointer"
							>
								{t("detail.overview.birthdayReminder")}
							</Label>
						</div>
					)}
					<div>
						<div className="flex items-center justify-between mb-1.5">
							<Label>{t("detail.overview.lastContact")}</Label>
							{lastContactAt && (
								<Button
									type="button"
									variant="neutral"
									size="sm"
									className="h-6 px-2 text-xs"
									onClick={() => setLastContactAt("")}
								>
									<X className="size-3" /> {t("detail.overview.clear")}
								</Button>
							)}
						</div>
						<Input
							type="datetime-local"
							value={lastContactAt}
							onChange={(e) => setLastContactAt(e.target.value)}
						/>
					</div>
					<div>
						<Label>{t("detail.overview.gender")}</Label>
						<RadioGroup
							value={gender}
							onValueChange={setGender}
							className="flex flex-wrap gap-4 mt-1"
						>
							<div className="flex items-center gap-2">
								<RadioGroupItem value="" id="edit-gender-unselected" />
								<Label
									htmlFor="edit-gender-unselected"
									className="font-normal cursor-pointer text-sub"
								>
									{t("detail.overview.genderUnselected")}
								</Label>
							</div>
							{genderOptions.map((opt) => (
								<div key={opt.value} className="flex items-center gap-2">
									<RadioGroupItem
										value={opt.value}
										id={`edit-gender-${opt.value}`}
									/>
									<Label
										htmlFor={`edit-gender-${opt.value}`}
										className="font-normal cursor-pointer"
									>
										{t(`gender.${opt.value}`)}
									</Label>
								</div>
							))}
						</RadioGroup>
					</div>
					<div>
						<Label>{t("detail.overview.bio")}</Label>
						<Textarea
							rows={3}
							value={bio}
							onChange={(e) => setBio(e.target.value)}
							placeholder={t("detail.overview.bioPlaceholder")}
						/>
					</div>
				</div>
			) : (
				<>
					<dl className="grid grid-cols-2 gap-x-4 gap-y-3 text-sm">
						{person.nickname && (
							<>
								<dt className="font-medium text-sub">
									{t("detail.overview.nickname")}
								</dt>
								<dd>{person.nickname}</dd>
							</>
						)}
						{person.is_self ? (
							<>
								<dt className="font-medium text-sub">
									{t("detail.overview.relationship")}
								</dt>
								<dd>
									<Badge variant="neutral">
										{t("detail.overview.selfProfile")}
									</Badge>
								</dd>
							</>
						) : null}
						{person.date_of_birth && (
							<>
								<dt className="font-medium text-sub">
									{t("detail.overview.dateOfBirth")}
								</dt>
								<dd className="flex items-center gap-2">
									{formatDate(person.date_of_birth)}
									{person.has_birthday_reminder && (
										<Badge variant="neutral">
											{t("detail.overview.reminderOn")}
										</Badge>
									)}
								</dd>
							</>
						)}
						{person.gender && (
							<>
								<dt className="font-medium text-sub">
									{t("detail.overview.gender")}
								</dt>
								<dd>
									{(() => {
										const known = genderOptions.find(
											(o) => o.value === person.gender,
										);
										return known ? t(`gender.${known.value}`) : person.gender;
									})()}
								</dd>
							</>
						)}
						{person.last_contact_at && (
							<>
								<dt className="font-medium text-sub">
									{t("detail.overview.lastContact")}
								</dt>
								<dd>{formatDate(person.last_contact_at)}</dd>
							</>
						)}
					</dl>
					{person.bio && (
						<p className="text-sm font-base whitespace-pre-wrap border-l-2 border-border pl-3">
							{person.bio}
						</p>
					)}
				</>
			)}
		</div>
	);
}

interface PersonDetailSectionsInnerProps {
	personId: number;
	onClose?: () => void;
}

function PersonDetailSectionsInner({
	personId,
	onClose,
}: PersonDetailSectionsInnerProps) {
	const { t } = useTranslation("people");
	const [editing, setEditing] = useState(false);
	const qc = useQueryClient();

	const { data: person } = useSuspenseQuery({
		queryKey: keys.people.detail(personId),
		queryFn: () => getPerson(personId),
	});

	const favoriteMutation = useMutation({
		mutationFn: () =>
			person.is_favorite ? unsetFavorite(personId) : setFavorite(personId),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: keys.people.detail(personId) });
			qc.invalidateQueries({ queryKey: keys.people.all });
		},
	});

	const [restoreError, setRestoreError] = useState<string | null>(null);

	const restoreMutation = useMutation({
		mutationFn: () => restorePerson(personId),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: keys.people.detail(personId) });
			qc.invalidateQueries({ queryKey: keys.people.all });
			qc.invalidateQueries({ queryKey: keys.search.all });
		},
		onError: (e) =>
			setRestoreError(apiErrorMessage(e, t("detail.restoreFailed"))),
	});

	const [confirmDeleteOpen, setConfirmDeleteOpen] = useState(false);
	const [deleteError, setDeleteError] = useState<string | null>(null);

	const deleteMutation = useMutation({
		mutationFn: () => deletePerson(personId),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: keys.people.detail(personId) });
			qc.invalidateQueries({ queryKey: keys.people.all });
			qc.invalidateQueries({ queryKey: keys.search.all });
			setConfirmDeleteOpen(false);
		},
		onError: (e) =>
			setDeleteError(apiErrorMessage(e, t("detail.deleteFailed"))),
	});

	const [confirmArchiveOpen, setConfirmArchiveOpen] = useState(false);
	const [archiveError, setArchiveError] = useState<string | null>(null);

	const archiveMutation = useMutation({
		mutationFn: () => archivePerson(personId),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: keys.people.detail(personId) });
			qc.invalidateQueries({ queryKey: keys.people.all });
			qc.invalidateQueries({ queryKey: keys.search.all });
			qc.invalidateQueries({ queryKey: keys.reminders.all });
			setConfirmArchiveOpen(false);
		},
		onError: (e) =>
			setArchiveError(apiErrorMessage(e, t("detail.archiveFailed"))),
	});

	const [confirmUnarchiveOpen, setConfirmUnarchiveOpen] = useState(false);
	const [unarchiveError, setUnarchiveError] = useState<string | null>(null);

	const unarchiveMutation = useMutation({
		mutationFn: () => unarchivePerson(personId),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: keys.people.detail(personId) });
			qc.invalidateQueries({ queryKey: keys.people.all });
			qc.invalidateQueries({ queryKey: keys.search.all });
			qc.invalidateQueries({ queryKey: keys.reminders.all });
			setConfirmUnarchiveOpen(false);
		},
		onError: (e) =>
			setUnarchiveError(apiErrorMessage(e, t("detail.unarchiveFailed"))),
	});

	const { data: settingsData } = useQuery({
		queryKey: ["settings"],
		queryFn: getSettings,
	});
	const retentionDays = settingsData?.people_delete_retention_days ?? 30;

	return (
		<div className="space-y-3">
			{person.deleted_at && (
				<Alert variant="warning">
					<AlertDescription className="flex flex-wrap items-center justify-between gap-2 w-full">
						<span>
							{retentionDays > 0
								? t("detail.pendingDeletion", {
										days: daysRemaining(person.deleted_at, retentionDays),
									})
								: t("detail.pendingDeletionNoPurge")}
						</span>
						{restoreError && (
							<span className="text-danger-fg">{restoreError}</span>
						)}
						<Button
							size="sm"
							disabled={restoreMutation.isPending}
							onClick={() => restoreMutation.mutate()}
						>
							{restoreMutation.isPending
								? t("detail.restoring")
								: t("detail.restore")}
						</Button>
					</AlertDescription>
				</Alert>
			)}
			{person.archived_at && (
				<Alert variant="warning">
					<AlertDescription className="flex flex-wrap items-center justify-between gap-2 w-full">
						<span>{t("detail.archivedBanner")}</span>
						{unarchiveError && (
							<span className="text-danger-fg">{unarchiveError}</span>
						)}
						<Button
							size="sm"
							onClick={() => {
								setUnarchiveError(null);
								setConfirmUnarchiveOpen(true);
							}}
						>
							{t("detail.unarchive")}
						</Button>
					</AlertDescription>
				</Alert>
			)}
			<div className="flex flex-wrap items-center justify-between gap-2">
				<div className="flex flex-wrap items-center gap-2">
					<h1 className="text-[18px] font-semibold tracking-tight text-ink font-display">
						{person.name}
					</h1>
					<button
						type="button"
						onClick={() => favoriteMutation.mutate()}
						disabled={favoriteMutation.isPending}
						aria-label={
							person.is_favorite ? t("detail.unfavorite") : t("detail.favorite")
						}
						className="text-sub hover:text-warning-fg disabled:opacity-50"
					>
						<Star
							className={`size-4 ${person.is_favorite ? "fill-warning-fg text-warning-fg" : ""}`}
						/>
					</button>
					{!editing && <QuickActions personId={person.id} />}
				</div>
				<div className="flex items-center gap-2">
					{!person.deleted_at && !person.is_self && !person.archived_at && (
						<Button
							variant="neutral"
							size="sm"
							onClick={() => {
								setArchiveError(null);
								setConfirmArchiveOpen(true);
							}}
						>
							{t("detail.archive")}
						</Button>
					)}
					{!person.deleted_at && !person.is_self && (
						<Button
							variant="neutral"
							size="sm"
							onClick={() => {
								setDeleteError(null);
								setConfirmDeleteOpen(true);
							}}
						>
							{t("detail.delete")}
						</Button>
					)}
					{onClose && (
						<Button variant="neutral" size="sm" onClick={onClose}>
							<X className="size-4" />
						</Button>
					)}
				</div>
			</div>

			<div className="grid grid-cols-1 gap-3 lg:grid-cols-2">
				<SectionCard>
					<OverviewSection
						person={person}
						editing={editing}
						onEdit={() => setEditing(true)}
						onCancel={() => setEditing(false)}
					/>
				</SectionCard>

				<div className="space-y-3">
					<SectionCard>
						<ContactsSection person={person} />
					</SectionCard>
					<SectionCard>
						<LabelsSection person={person} />
					</SectionCard>
				</div>

				<SectionCard>
					<RelationshipsSection
						personId={personId}
						personName={formatPersonName(person.name, person.nickname)}
					/>
				</SectionCard>
				<SectionCard>
					<ImportantDatesSection personId={personId} person={person} />
				</SectionCard>

				<SectionCard>
					<JournalSection personId={personId} />
				</SectionCard>
				<SectionCard>
					<GiftsSection personId={personId} />
				</SectionCard>
				<SectionCard>
					<NotesSection personId={personId} />
				</SectionCard>

				<SectionCard>
					<LocationsSection person={person} />
				</SectionCard>
				<SectionCard>
					<WorkHistorySection personId={personId} />
				</SectionCard>

				<div className="lg:col-span-2">
					<SectionCard>
						<PersonSectionRelationshipGraph personId={personId} />
					</SectionCard>
				</div>
			</div>

			<Dialog
				open={confirmDeleteOpen}
				onOpenChange={(v) => !v && setConfirmDeleteOpen(false)}
			>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>
							{t("detail.deleteDialog.title", { name: person.name })}
						</DialogTitle>
					</DialogHeader>
					{deleteError && (
						<Alert variant="destructive">
							<AlertDescription>{deleteError}</AlertDescription>
						</Alert>
					)}
					<p className="text-[13px] text-sub">
						{retentionDays > 0
							? t("detail.deleteDialog.body", {
									name: person.name,
									days: retentionDays,
								})
							: t("detail.deleteDialog.bodyNoPurge", { name: person.name })}
					</p>
					<DialogFooter>
						<Button
							variant="neutral"
							onClick={() => setConfirmDeleteOpen(false)}
						>
							{t("detail.deleteDialog.cancel")}
						</Button>
						<Button
							variant="destructive"
							disabled={deleteMutation.isPending}
							onClick={() => deleteMutation.mutate()}
						>
							{deleteMutation.isPending
								? t("detail.deleteDialog.deleting")
								: t("detail.deleteDialog.delete")}
						</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>

			<Dialog
				open={confirmArchiveOpen}
				onOpenChange={(v) => !v && setConfirmArchiveOpen(false)}
			>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>
							{t("detail.archiveDialog.title", { name: person.name })}
						</DialogTitle>
					</DialogHeader>
					{archiveError && (
						<Alert variant="destructive">
							<AlertDescription>{archiveError}</AlertDescription>
						</Alert>
					)}
					<p className="text-[13px] text-sub">
						{t("detail.archiveDialog.body", { name: person.name })}
					</p>
					<DialogFooter>
						<Button
							variant="neutral"
							onClick={() => setConfirmArchiveOpen(false)}
						>
							{t("detail.archiveDialog.cancel")}
						</Button>
						<Button
							disabled={archiveMutation.isPending}
							onClick={() => archiveMutation.mutate()}
						>
							{archiveMutation.isPending
								? t("detail.archiveDialog.archiving")
								: t("detail.archiveDialog.archive")}
						</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>

			<Dialog
				open={confirmUnarchiveOpen}
				onOpenChange={(v) => !v && setConfirmUnarchiveOpen(false)}
			>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>
							{t("detail.unarchiveDialog.title", { name: person.name })}
						</DialogTitle>
					</DialogHeader>
					{unarchiveError && (
						<Alert variant="destructive">
							<AlertDescription>{unarchiveError}</AlertDescription>
						</Alert>
					)}
					<p className="text-[13px] text-sub">
						{t("detail.unarchiveDialog.body", { name: person.name })}
					</p>
					<DialogFooter>
						<Button
							variant="neutral"
							onClick={() => setConfirmUnarchiveOpen(false)}
						>
							{t("detail.unarchiveDialog.cancel")}
						</Button>
						<Button
							disabled={unarchiveMutation.isPending}
							onClick={() => unarchiveMutation.mutate()}
						>
							{unarchiveMutation.isPending
								? t("detail.unarchiveDialog.unarchiving")
								: t("detail.unarchiveDialog.unarchive")}
						</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>
		</div>
	);
}

function PersonDetailFallback() {
	const { t } = useTranslation("people");
	return (
		<div className="py-12 text-center text-sub font-base">
			{t("detail.loading")}
		</div>
	);
}

interface PersonDetailSectionsProps {
	personId: number;
	onClose?: () => void;
}

export function PersonDetailSections({
	personId,
	onClose,
}: PersonDetailSectionsProps) {
	return (
		<QueryBoundary fallback={<PersonDetailFallback />}>
			<PersonDetailSectionsInner
				key={personId}
				personId={personId}
				onClose={onClose}
			/>
		</QueryBoundary>
	);
}
