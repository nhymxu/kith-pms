// Journal create/edit form — TanStack Form + Zod, people + label multi-select
import { useForm } from "@tanstack/react-form";
import { useQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { Plus, X } from "lucide-react";
import { useState } from "react";
import { FormField } from "#/components/form/form-field";
import { SubmitButton } from "#/components/form/submit-button";
import { Alert, AlertDescription } from "#/components/ui/alert";
import { Badge } from "#/components/ui/badge";
import { Label } from "#/components/ui/label";
import { Textarea } from "#/components/ui/textarea";
import { listJournalLabels } from "#/endpoints/journal-labels";
import { getMe } from "#/endpoints/me";
import { getAvatarUrl, listPeople } from "#/endpoints/people";
import { keys } from "#/query-keys";
import {
	type JournalActivity,
	type JournalRequest,
	journalRequestSchema,
} from "#/schemas/journal";

interface JournalFormProps {
	initial?: Partial<JournalActivity>;
	onSubmit: (values: JournalRequest) => Promise<void>;
	submitLabel?: string;
	defaultPersonId?: number;
}

export function JournalForm({
	initial,
	onSubmit,
	submitLabel = "Save Entry",
	defaultPersonId,
}: JournalFormProps) {
	const [apiError, setApiError] = useState<string | null>(null);
	const [searchQ, setSearchQ] = useState("");

	const { data: peopleList } = useQuery({
		queryKey: keys.people.list({ q: searchQ || undefined }),
		queryFn: () => listPeople({ q: searchQ || undefined, page_size: 50 }),
	});

	const { data: allLabels } = useQuery({
		queryKey: keys.journalLabels.list(),
		queryFn: listJournalLabels,
	});
	const { data: myProfile } = useQuery({
		queryKey: keys.me.profile(),
		queryFn: getMe,
	});

	const form = useForm({
		defaultValues: {
			title: initial?.title ?? "",
			content: initial?.content ?? "",
			occurred_at_date:
				initial?.occurred_at_date ?? new Date().toISOString().slice(0, 10),
			occurred_at_time: initial?.occurred_at_time ?? "",
			person_ids:
				initial?.people?.map((p) => p.person_id) ??
				(defaultPersonId ? [defaultPersonId] : []),
			label_ids: initial?.labels?.map((l) => l.id) ?? [],
		} satisfies JournalRequest,
		validators: {
			onSubmit: ({ value }) => {
				const result = journalRequestSchema.safeParse(value);
				if (!result.success)
					return result.error.issues.map((i) => i.message).join(", ");
				return undefined;
			},
		},
		onSubmit: async ({ value }) => {
			setApiError(null);
			try {
				await onSubmit(value as JournalRequest);
			} catch (err) {
				setApiError(
					err instanceof Error ? err.message : "Failed to save entry",
				);
			}
		},
	});

	return (
		<form
			onSubmit={(e) => {
				e.preventDefault();
				form.handleSubmit();
			}}
			className="space-y-4 max-w-2xl"
		>
			{apiError && (
				<Alert variant="destructive">
					<AlertDescription>{apiError}</AlertDescription>
				</Alert>
			)}

			<form.Field name="title">
				{(f) => (
					<FormField field={f} label="Title *" placeholder="What happened?" />
				)}
			</form.Field>

			<div className="grid grid-cols-2 gap-4">
				<form.Field name="occurred_at_date">
					{(f) => <FormField field={f} label="Date *" type="date" />}
				</form.Field>
				<form.Field name="occurred_at_time">
					{(f) => <FormField field={f} label="Time (optional)" type="time" />}
				</form.Field>
			</div>

			{/* People multi-select */}
			<form.Field name="person_ids">
				{(f) => {
					const selectedIds: number[] = Array.isArray(f.state.value)
						? f.state.value
						: [];
					const selectedPeople =
						peopleList?.items.filter((p) => selectedIds.includes(p.id)) ?? [];
					const unselected =
						peopleList?.items.filter((p) => !selectedIds.includes(p.id)) ?? [];

					return (
						<div className="space-y-2">
							<div className="flex items-center gap-2">
								<Label>People</Label>
								{myProfile && !selectedIds.includes(myProfile.id) && (
									<button
										type="button"
										onClick={() =>
											f.handleChange([...selectedIds, myProfile.id])
										}
										className="text-xs text-indigo-600 hover:text-indigo-800 font-medium"
									>
										+ Me
									</button>
								)}
							</div>
							{selectedPeople.length > 0 && (
								<div className="flex flex-wrap gap-1">
									{selectedPeople.map((p) => (
										<span
											key={p.id}
											className="flex items-center gap-1.5 rounded-full border border-zinc-200 bg-white px-2 py-0.5"
										>
											<span className="size-5 rounded-full overflow-hidden shrink-0 bg-zinc-100 flex items-center justify-center text-[9px] font-medium text-zinc-600">
												{p.avatar_path ? (
													<img
														src={getAvatarUrl(p.id)}
														alt={p.name}
														className="size-full object-cover"
													/>
												) : (
													p.name.charAt(0).toUpperCase()
												)}
											</span>
											<span className="text-[11px] text-zinc-700 leading-none">
												{p.nickname || p.name}
											</span>
											<button
												type="button"
												onClick={() =>
													f.handleChange(
														selectedIds.filter((id) => id !== p.id),
													)
												}
												className="ml-0.5 hover:text-destructive"
											>
												<X className="size-3" />
											</button>
										</span>
									))}
								</div>
							)}
							<div className="space-y-1">
								<input
									type="text"
									value={searchQ}
									onChange={(e) => setSearchQ(e.target.value)}
									placeholder="Search people to add…"
									className="h-9 w-full border border-zinc-200 rounded-md bg-white px-3 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
								/>
								{unselected.length > 0 && (
									<div className="flex flex-wrap gap-1 pt-1">
										{unselected.slice(0, 10).map((p) => (
											<button
												key={p.id}
												type="button"
												onClick={() => f.handleChange([...selectedIds, p.id])}
												className="flex items-center gap-1.5 rounded-full border border-dashed border-zinc-300 bg-white px-2 py-0.5 hover:border-main transition-colors"
											>
												<span className="size-5 rounded-full overflow-hidden shrink-0 bg-zinc-100 flex items-center justify-center text-[9px] font-medium text-zinc-600">
													{p.avatar_path ? (
														<img
															src={getAvatarUrl(p.id)}
															alt={p.name}
															className="size-full object-cover"
														/>
													) : (
														p.name.charAt(0).toUpperCase()
													)}
												</span>
												<span className="text-[11px] text-zinc-700 leading-none">
													{p.nickname || p.name}
												</span>
												<Plus className="size-3 text-zinc-400" />
											</button>
										))}
									</div>
								)}
							</div>
						</div>
					);
				}}
			</form.Field>

			{/* Labels multi-select */}
			<form.Field name="label_ids">
				{(f) => {
					const selectedIds: number[] = Array.isArray(f.state.value)
						? f.state.value
						: [];
					const selected =
						allLabels?.filter((l) => selectedIds.includes(l.id)) ?? [];
					const unselected =
						allLabels?.filter((l) => !selectedIds.includes(l.id)) ?? [];

					return (
						<div className="space-y-2">
							<Label>Labels</Label>
							{selected.length > 0 && (
								<div className="flex flex-wrap gap-1">
									{selected.map((l) => (
										<Badge
											key={l.id}
											variant="neutral"
											className="flex items-center gap-1"
											style={{ borderColor: l.color }}
										>
											{l.name}
											<button
												type="button"
												onClick={() =>
													f.handleChange(
														selectedIds.filter((id) => id !== l.id),
													)
												}
												className="ml-0.5 hover:text-destructive"
											>
												<X className="size-3" />
											</button>
										</Badge>
									))}
								</div>
							)}
							{allLabels && allLabels.length > 0 ? (
								unselected.length > 0 && (
									<div className="flex flex-wrap gap-1">
										{unselected.map((l) => (
											<button
												key={l.id}
												type="button"
												onClick={() => f.handleChange([...selectedIds, l.id])}
												className="flex items-center gap-1 text-xs border border-dashed border-zinc-300 rounded-md px-2 py-1 hover:border-main transition-colors"
											>
												<Plus className="size-3" />
												{l.name}
											</button>
										))}
									</div>
								)
							) : (
								<p className="text-[12px] text-zinc-400">
									No labels yet. Create some in{" "}
									<Link
										to="/settings/journal-labels"
										className="text-indigo-600 hover:underline"
									>
										Settings → Journal Labels
									</Link>
								</p>
							)}
						</div>
					);
				}}
			</form.Field>

			{/* Content */}
			<div className="space-y-1.5">
				<Label>Content</Label>
				<form.Field name="content">
					{(f) => (
						<Textarea
							value={f.state.value}
							onBlur={f.handleBlur}
							onChange={(e) => f.handleChange(e.target.value)}
							placeholder="Write about the interaction…"
							rows={6}
						/>
					)}
				</form.Field>
			</div>

			<form.Subscribe selector={(s) => s.isSubmitting}>
				{(isSubmitting) => (
					<SubmitButton isPending={isSubmitting} pendingLabel="Saving…">
						{submitLabel}
					</SubmitButton>
				)}
			</form.Subscribe>
		</form>
	);
}
