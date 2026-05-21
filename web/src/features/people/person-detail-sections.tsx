import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Pencil, X } from "lucide-react";
import { useState } from "react";
import { Badge } from "#/components/ui/badge";
import { Button } from "#/components/ui/button";
import { Input } from "#/components/ui/input";
import { Label } from "#/components/ui/label";
import { RadioGroup, RadioGroupItem } from "#/components/ui/radio-group";
import { Textarea } from "#/components/ui/textarea";
import { getPerson, updatePerson } from "#/endpoints/people";
import { formatDate } from "#/lib/format-datetime";
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
import { RelationshipsSection } from "./person-section-relationships";
import { WorkHistorySection } from "./person-section-work-history";
import { QuickActions } from "./quick-actions";

function SectionCard({ children }: { children: React.ReactNode }) {
	return (
		<div className="rounded-lg border border-zinc-200 bg-white p-4">
			{children}
		</div>
	);
}

function SectionHeading({ children }: { children: React.ReactNode }) {
	return (
		<h2 className="text-[11px] font-semibold uppercase tracking-widest text-zinc-400 mb-2">
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
	const qc = useQueryClient();
	const [nickname, setNickname] = useState(person.nickname);
	const [gender, setGender] = useState(person.gender ?? "");
	const [relationshipType, setRelationshipType] = useState(
		person.relationship_type,
	);
	const [dob, setDob] = useState(person.date_of_birth ?? "");
	const [notes, setNotes] = useState(person.other_notes);

	const saveMutation = useMutation({
		mutationFn: () =>
			updatePerson(person.id, {
				name: person.name,
				nickname,
				gender,
				relationship_type: relationshipType,
				date_of_birth: dob,
				other_notes: notes,
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
				<SectionHeading>Overview</SectionHeading>
				{editing ? (
					<div className="flex gap-2">
						<Button variant="neutral" size="sm" onClick={onCancel}>
							Cancel
						</Button>
						<Button
							size="sm"
							disabled={saveMutation.isPending}
							onClick={() => saveMutation.mutate()}
						>
							{saveMutation.isPending ? "Saving…" : "Save"}
						</Button>
					</div>
				) : (
					<Button variant="neutral" size="sm" onClick={onEdit}>
						<Pencil className="size-3" /> Edit
					</Button>
				)}
			</div>

			<AvatarUploader
				personId={person.id}
				hasAvatar={Boolean(person.avatar_path)}
				showControls={editing}
			/>

			{!editing && <QuickActions personId={person.id} />}

			{editing ? (
				<div className="space-y-3">
					<div>
						<Label>Nickname</Label>
						<Input
							value={nickname}
							onChange={(e) => setNickname(e.target.value)}
							placeholder="Nickname"
						/>
					</div>
					{!person.is_self && (
						<div>
							<Label>Relationship type</Label>
							<Input
								value={relationshipType}
								onChange={(e) => setRelationshipType(e.target.value)}
								placeholder="e.g. Friend, Colleague"
							/>
						</div>
					)}
					<div>
						<Label>Date of birth</Label>
						<Input
							type="date"
							value={dob}
							onChange={(e) => setDob(e.target.value)}
						/>
					</div>
					<div>
						<Label>Gender</Label>
						<RadioGroup
							value={gender}
							onValueChange={setGender}
							className="flex flex-wrap gap-4 mt-1"
						>
							<div className="flex items-center gap-2">
								<RadioGroupItem value="" id="edit-gender-unselected" />
								<Label
									htmlFor="edit-gender-unselected"
									className="font-normal cursor-pointer text-zinc-400"
								>
									Unselected
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
										{opt.label}
									</Label>
								</div>
							))}
						</RadioGroup>
					</div>
					<div>
						<Label>Notes</Label>
						<Textarea
							rows={3}
							value={notes}
							onChange={(e) => setNotes(e.target.value)}
							placeholder="Notes…"
						/>
					</div>
				</div>
			) : (
				<>
					<dl className="grid grid-cols-2 gap-x-4 gap-y-3 text-sm">
						{person.nickname && (
							<>
								<dt className="font-medium text-zinc-500">Nickname</dt>
								<dd>{person.nickname}</dd>
							</>
						)}
						{person.is_self ? (
							<>
								<dt className="font-medium text-zinc-500">Relationship</dt>
								<dd>
									<Badge variant="neutral">Self profile</Badge>
								</dd>
							</>
						) : person.relationship_type ? (
							<>
								<dt className="font-medium text-zinc-500">Relationship</dt>
								<dd>{person.relationship_type}</dd>
							</>
						) : null}
						{person.date_of_birth && (
							<>
								<dt className="font-medium text-zinc-500">Date of birth</dt>
								<dd>{person.date_of_birth}</dd>
							</>
						)}
						{person.gender && (
							<>
								<dt className="font-medium text-zinc-500">Gender</dt>
								<dd>
									{genderOptions.find((o) => o.value === person.gender)
										?.label ?? person.gender}
								</dd>
							</>
						)}
						{person.last_contact_at && (
							<>
								<dt className="font-medium text-zinc-500">Last contact</dt>
								<dd>{formatDate(person.last_contact_at)}</dd>
							</>
						)}
					</dl>
					{person.other_notes && (
						<p className="text-sm font-base whitespace-pre-wrap border-l-2 border-border pl-3">
							{person.other_notes}
						</p>
					)}
				</>
			)}
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
	const [editing, setEditing] = useState(false);

	const {
		data: person,
		isLoading,
		error,
	} = useQuery({
		queryKey: keys.people.detail(personId),
		queryFn: () => getPerson(personId),
	});

	if (isLoading)
		return (
			<div className="py-12 text-center text-zinc-400 font-base">Loading…</div>
		);
	if (error || !person)
		return (
			<div className="py-12 text-center text-zinc-400 font-base">
				Person not found.
			</div>
		);

	return (
		<div className="space-y-3">
			<div className="flex items-center justify-between">
				<h1 className="text-[18px] font-semibold tracking-tight text-zinc-900">
					{person.name}
				</h1>
				{onClose && (
					<Button variant="neutral" size="sm" onClick={onClose}>
						<X className="size-4" />
					</Button>
				)}
			</div>

			<SectionCard>
				<OverviewSection
					person={person}
					editing={editing}
					onEdit={() => setEditing(true)}
					onCancel={() => setEditing(false)}
				/>
			</SectionCard>

			<SectionCard>
				<ContactsSection person={person} />
			</SectionCard>
			<SectionCard>
				<LocationsSection person={person} />
			</SectionCard>
			<SectionCard>
				<LabelsSection person={person} />
			</SectionCard>
			<SectionCard>
				<RelationshipsSection personId={personId} />
			</SectionCard>
			<SectionCard>
				<JournalSection personId={personId} />
			</SectionCard>
			<SectionCard>
				<WorkHistorySection personId={personId} />
			</SectionCard>
			<SectionCard>
				<ImportantDatesSection personId={personId} person={person} />
			</SectionCard>
			<SectionCard>
				<GiftsSection personId={personId} />
			</SectionCard>
		</div>
	);
}
