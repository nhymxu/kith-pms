import {
	keepPreviousData,
	useMutation,
	useQuery,
	useQueryClient,
} from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import type { ColumnDef, RowSelectionState } from "@tanstack/react-table";
import { Star } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { DataTable } from "#/components/data-table/data-table";
import { Badge } from "#/components/ui/badge";
import { Button } from "#/components/ui/button";
import { Input } from "#/components/ui/input";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "#/components/ui/select";
import {
	getAvatarUrl,
	listPeople,
	setFavorite,
	unsetFavorite,
} from "#/endpoints/people";
import { listPeopleLabels } from "#/endpoints/people-labels";
import { formatDate } from "#/lib/format-datetime";
import { keys } from "#/query-keys";
import type { Person } from "#/schemas/person";
import { BulkActionBar } from "./bulk-action-bar";

const SORT_OPTIONS = [
	{ value: "name", label: "Name A→Z" },
	{ value: "-name", label: "Name Z→A" },
	{ value: "-last_contact", label: "Last contact: newest" },
	{ value: "last_contact", label: "Last contact: oldest" },
	{ value: "-favorite", label: "Favorites first" },
] as const;

type SortValue = (typeof SORT_OPTIONS)[number]["value"];

function useDebounce<T>(value: T, ms = 300): T {
	const [debounced, setDebounced] = useState(value);
	useEffect(() => {
		const t = setTimeout(() => setDebounced(value), ms);
		return () => clearTimeout(t);
	}, [value, ms]);
	return debounced;
}

interface PeopleTableProps {
	q?: string;
	labels?: number[];
	page?: number;
	page_size?: number;
	sort?: string;
	favoriteOnly?: boolean;
	onSearchChange: (q: string) => void;
	onLabelsChange: (labels: number[]) => void;
	onPageChange: (page: number) => void;
	onSortChange: (sort: SortValue) => void;
	onFavoriteOnlyChange: (v: boolean) => void;
}

function buildColumns(
	favoriteMutation: ReturnType<
		typeof useMutation<void, Error, { id: number; favorite: boolean }>
	>,
): ColumnDef<Person>[] {
	return [
		{
			id: "favorite",
			header: "",
			size: 36,
			cell: ({ row }) => {
				const p = row.original;
				return (
					<button
						type="button"
						onClick={(e) => {
							e.preventDefault();
							e.stopPropagation();
							favoriteMutation.mutate({ id: p.id, favorite: !p.is_favorite });
						}}
						className="text-zinc-300 hover:text-amber-500 disabled:opacity-50"
						disabled={favoriteMutation.isPending}
						aria-label={p.is_favorite ? "Unfavorite" : "Favorite"}
					>
						<Star
							className={`size-4 ${p.is_favorite ? "fill-amber-400 text-amber-500" : ""}`}
						/>
					</button>
				);
			},
		},
		{
			id: "name",
			header: "Name",
			accessorKey: "name",
			cell: ({ row }) => {
				const p = row.original;
				const hasAvatar = Boolean(p.avatar_path);
				return (
					<Link
						to="/people/$personId"
						params={{ personId: String(p.id) }}
						className="flex items-center gap-2 hover:underline"
					>
						<div className="size-7 rounded-full overflow-hidden shrink-0 bg-zinc-100 flex items-center justify-center text-[11px] font-medium text-zinc-700 font-mono">
							{hasAvatar ? (
								<img
									src={getAvatarUrl(p.id)}
									alt={p.name}
									className="size-full object-cover"
								/>
							) : (
								<span>{p.name.charAt(0).toUpperCase()}</span>
							)}
						</div>
						<div>
							<p className="text-[13px] text-zinc-900">{p.name}</p>
							{p.nickname && (
								<p className="text-[11px] text-zinc-500">"{p.nickname}"</p>
							)}
						</div>
					</Link>
				);
			},
		},
		{
			id: "labels",
			header: "Labels",
			cell: ({ row }) => {
				const labels = row.original.labels ?? [];
				return (
					<div className="flex flex-wrap gap-1">
						{labels.slice(0, 3).map((l) => (
							<Badge
								key={l.id}
								variant="neutral"
								style={{ borderColor: l.color }}
							>
								{l.name}
							</Badge>
						))}
						{labels.length > 3 && (
							<Badge variant="neutral">+{labels.length - 3}</Badge>
						)}
					</div>
				);
			},
		},
		{
			id: "last_contact_at",
			header: "Last contact",
			accessorKey: "last_contact_at",
			cell: ({ getValue }) => {
				const v = getValue<string | null>();
				return v ? (
					<span className="font-mono text-[12px] text-zinc-500">
						{formatDate(v)}
					</span>
				) : (
					<span className="text-[12px] text-zinc-300">—</span>
				);
			},
		},
		{
			id: "actions",
			header: "",
			size: 80,
			cell: ({ row }) => (
				<Button variant="ghost" size="sm" asChild>
					<Link
						to="/people/$personId/edit"
						params={{ personId: String(row.original.id) }}
					>
						Edit
					</Link>
				</Button>
			),
		},
	];
}

export function PeopleTable({
	q = "",
	labels = [],
	page = 1,
	page_size = 20,
	sort = "name",
	favoriteOnly = false,
	onSearchChange,
	onLabelsChange,
	onPageChange,
	onSortChange,
	onFavoriteOnlyChange,
}: PeopleTableProps) {
	const [localQ, setLocalQ] = useState(q);
	const debouncedQ = useDebounce(localQ, 300);
	const isFirst = useRef(true);
	const [rowSelection, setRowSelection] = useState<RowSelectionState>({});
	const qc = useQueryClient();
	const favoriteMutation = useMutation({
		mutationFn: ({ id, favorite }: { id: number; favorite: boolean }) =>
			favorite ? setFavorite(id) : unsetFavorite(id),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: keys.people.all });
		},
	});
	const columns = buildColumns(favoriteMutation);

	useEffect(() => {
		if (isFirst.current) {
			isFirst.current = false;
			return;
		}
		onSearchChange(debouncedQ);
	}, [debouncedQ, onSearchChange]);

	useEffect(() => {
		setLocalQ(q);
	}, [q]);

	const { data, isLoading } = useQuery({
		queryKey: keys.people.list({
			q: debouncedQ || undefined,
			labels: labels.length ? labels : undefined,
			page,
			page_size,
			sort,
			favorite_only: favoriteOnly || undefined,
		}),
		queryFn: () =>
			listPeople({
				q: debouncedQ || undefined,
				labels: labels.length ? labels : undefined,
				page,
				page_size,
				sort,
				favorite_only: favoriteOnly || undefined,
			}),
		placeholderData: keepPreviousData,
	});

	const { data: allLabelsData } = useQuery({
		queryKey: keys.peopleLabels.list(),
		queryFn: listPeopleLabels,
	});

	const rows = data?.items ?? [];
	const selectedIDs = Object.keys(rowSelection).map(Number);

	return (
		<div className="space-y-3">
			<div className="flex items-center gap-3">
				<Input
					placeholder="Search people…"
					value={localQ}
					onChange={(e) => setLocalQ(e.target.value)}
					className="max-w-xs"
				/>
				<Select
					value={sort}
					onValueChange={(v) => onSortChange(v as SortValue)}
				>
					<SelectTrigger className="w-48">
						<SelectValue />
					</SelectTrigger>
					<SelectContent>
						{SORT_OPTIONS.map((opt) => (
							<SelectItem key={opt.value} value={opt.value}>
								{opt.label}
							</SelectItem>
						))}
					</SelectContent>
				</Select>
				<button
					type="button"
					onClick={() => onFavoriteOnlyChange(!favoriteOnly)}
					className={`text-xs border rounded-md px-2 py-1 transition-colors flex items-center gap-1 ${favoriteOnly ? "border-main bg-main/10" : "border-zinc-200 hover:border-zinc-400"}`}
				>
					<Star
						className={`size-3 ${favoriteOnly ? "fill-amber-400 text-amber-500" : ""}`}
					/>
					Favorites only
				</button>
			</div>
			{allLabelsData && allLabelsData.length > 0 && (
				<div className="flex flex-wrap gap-2">
					{allLabelsData.map((l) => {
						const active = labels.includes(l.id);
						return (
							<button
								key={l.id}
								type="button"
								onClick={() => {
									const next = active
										? labels.filter((id) => id !== l.id)
										: [...labels, l.id];
									onLabelsChange(next);
								}}
								className={`text-xs border rounded-md px-2 py-1 transition-colors ${active ? "border-main bg-main/10" : "border-zinc-200 hover:border-zinc-400"}`}
								style={active ? { borderColor: l.color } : undefined}
							>
								{l.name}
							</button>
						);
					})}
					{labels.length > 0 && (
						<button
							type="button"
							onClick={() => onLabelsChange([])}
							className="text-xs text-zinc-400 hover:text-zinc-700"
						>
							Clear
						</button>
					)}
				</div>
			)}
			<DataTable
				columns={columns}
				data={rows}
				pageSize={page_size}
				totalCount={data?.total}
				pageIndex={page - 1}
				onPageChange={(idx) => onPageChange(idx + 1)}
				hideToolbar
				enableRowSelection
				rowSelection={rowSelection}
				onRowSelectionChange={setRowSelection}
				getRowId={(row) => String(row.id)}
				emptyState={
					isLoading ? (
						<span className="text-sm font-base text-foreground/50">
							Loading…
						</span>
					) : (
						<span className="text-sm font-base text-foreground/50">
							No people found.
						</span>
					)
				}
			/>
			{selectedIDs.length > 0 && (
				<BulkActionBar
					selectedCount={selectedIDs.length}
					personIds={selectedIDs}
					onClear={() => setRowSelection({})}
				/>
			)}
		</div>
	);
}
