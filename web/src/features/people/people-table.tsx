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
import { useDebounce } from "#/hooks/use-debounce";
import { formatDate } from "#/lib/format-datetime";
import { keys } from "#/query-keys";
import type { Person } from "#/schemas/person";
import { BulkActionBar } from "./bulk-action-bar";

const SORT_OPTIONS = [
	{ value: "name", label: "Name A→Z" },
	{ value: "-name", label: "Name Z→A" },
	{ value: "-last_contact", label: "Last contact: newest" },
	{ value: "last_contact", label: "Last contact: oldest" },
] as const;

type SortValue = (typeof SORT_OPTIONS)[number]["value"];

interface PeopleTableProps {
	q?: string;
	labels?: number[];
	page?: number;
	page_size?: number;
	sort?: string;
	favoriteOnly?: boolean;
	favoriteFirst?: boolean;
	allowToggle?: boolean;
	pageSizeSelector?: React.ReactNode;
	onSearchChange: (q: string) => void;
	onLabelsChange: (labels: number[]) => void;
	onPageChange: (page: number) => void;
	onSortChange: (sort: SortValue) => void;
	onFavoriteOnlyChange: (v: boolean) => void;
	onFavoriteFirstChange: (v: boolean) => void;
}

function buildColumns(
	favoriteMutation: ReturnType<
		typeof useMutation<void, Error, { id: number; favorite: boolean }>
	>,
	allowToggle: boolean,
): ColumnDef<Person>[] {
	return [
		{
			id: "favorite",
			header: "",
			size: 36,
			cell: ({ row }) => {
				const p = row.original;
				if (!allowToggle) {
					return (
						<Star
							className={`size-4 ${p.is_favorite ? "fill-warning-fg text-warning-fg" : "text-sub"}`}
							aria-label={p.is_favorite ? "Favorite" : undefined}
						/>
					);
				}
				return (
					<button
						type="button"
						onClick={(e) => {
							e.preventDefault();
							e.stopPropagation();
							favoriteMutation.mutate({ id: p.id, favorite: !p.is_favorite });
						}}
						className="text-sub hover:text-warning-fg disabled:opacity-50"
						disabled={favoriteMutation.isPending}
						aria-label={p.is_favorite ? "Unfavorite" : "Favorite"}
					>
						<Star
							className={`size-4 ${p.is_favorite ? "fill-warning-fg text-warning-fg" : ""}`}
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
						<div className="size-7 rounded-full overflow-hidden shrink-0 bg-chip flex items-center justify-center text-[11px] font-medium text-chip-fg font-mono">
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
							<p className="text-[13px] text-ink">{p.name}</p>
							{p.nickname && (
								<p className="text-[11px] text-sub">"{p.nickname}"</p>
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
					<span className="font-mono text-[12px] text-sub">
						{formatDate(v)}
					</span>
				) : (
					<span className="text-[12px] text-sub">—</span>
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
	favoriteFirst = false,
	allowToggle = true,
	pageSizeSelector,
	onSearchChange,
	onLabelsChange,
	onPageChange,
	onSortChange,
	onFavoriteOnlyChange,
	onFavoriteFirstChange,
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
	const columns = buildColumns(favoriteMutation, allowToggle);

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
			favorite_first: favoriteFirst || undefined,
		}),
		queryFn: () =>
			listPeople({
				q: debouncedQ || undefined,
				labels: labels.length ? labels : undefined,
				page,
				page_size,
				sort,
				favorite_only: favoriteOnly || undefined,
				favorite_first: favoriteFirst || undefined,
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
					className={`h-9 text-xs border rounded-md px-3 transition-colors flex items-center gap-1 ${favoriteOnly ? "border-main bg-main/10" : "border-line hover:border-sub"}`}
				>
					<Star
						className={`size-3 ${favoriteOnly ? "fill-warning-fg text-warning-fg" : ""}`}
					/>
					Favorites only
				</button>
				<button
					type="button"
					onClick={() => onFavoriteFirstChange(!favoriteFirst)}
					className={`h-9 text-xs border rounded-md px-3 transition-colors flex items-center gap-1 ${favoriteFirst ? "border-main bg-main/10" : "border-line hover:border-sub"}`}
				>
					<Star
						className={`size-3 ${favoriteFirst ? "fill-warning-fg text-warning-fg" : ""}`}
					/>
					Favorites first
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
								className="text-xs border rounded-full px-3 py-1 transition-colors text-ink"
								style={{
									borderColor: l.color,
									backgroundColor: active ? `${l.color}26` : undefined,
								}}
							>
								{l.name}
							</button>
						);
					})}
					{labels.length > 0 && (
						<button
							type="button"
							onClick={() => onLabelsChange([])}
							className="text-xs text-sub hover:text-ink"
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
				pageSizeSelector={pageSizeSelector}
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
