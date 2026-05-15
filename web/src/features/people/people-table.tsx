import { useQuery } from "@tanstack/react-query"
import { Link } from "@tanstack/react-router"
import { useEffect, useRef, useState } from "react"
import type { ColumnDef } from "@tanstack/react-table"
import { DataTable } from "#/components/data-table/data-table"
import { Badge } from "#/components/ui/badge"
import { Input } from "#/components/ui/input"
import { Button } from "#/components/ui/button"
import { keys } from "#/query-keys"
import { listPeople, getAvatarUrl } from "#/endpoints/people"
import { listLabels } from "#/endpoints/labels"
import type { Person } from "#/schemas/person"

function useDebounce<T>(value: T, ms = 300): T {
	const [debounced, setDebounced] = useState(value)
	useEffect(() => {
		const t = setTimeout(() => setDebounced(value), ms)
		return () => clearTimeout(t)
	}, [value, ms])
	return debounced
}

interface PeopleTableProps {
	q?: string
	labels?: number[]
	page?: number
	page_size?: number
	onSearchChange: (q: string) => void
}

const columns: ColumnDef<Person>[] = [
	{
		id: "avatar",
		header: "",
		size: 48,
		cell: ({ row }) => {
			const p = row.original
			const hasAvatar = Boolean(p.avatar_path)
			return (
				<div className="size-8 rounded-base border-2 border-border overflow-hidden shrink-0 bg-secondary-background flex items-center justify-center text-xs font-heading">
					{hasAvatar ? (
						<img src={getAvatarUrl(p.id)} alt={p.name} className="size-full object-cover" />
					) : (
						<span>{p.name.charAt(0).toUpperCase()}</span>
					)}
				</div>
			)
		},
	},
	{
		id: "name",
		header: "Name",
		accessorKey: "name",
		cell: ({ row }) => {
			const p = row.original
			return (
				<Link
					to="/people/$personId"
					params={{ personId: String(p.id) }}
					className="block hover:underline"
				>
					<p className="font-heading text-sm">{p.name}</p>
					{p.nickname && <p className="text-xs font-base text-foreground/50">"{p.nickname}"</p>}
				</Link>
			)
		},
	},
	{
		id: "labels",
		header: "Labels",
		cell: ({ row }) => {
			const labels = row.original.labels ?? []
			return (
				<div className="flex flex-wrap gap-1">
					{labels.slice(0, 3).map((l) => (
						<Badge key={l.id} variant="neutral" style={{ borderColor: l.color }}>
							{l.name}
						</Badge>
					))}
					{labels.length > 3 && (
						<Badge variant="neutral">+{labels.length - 3}</Badge>
					)}
				</div>
			)
		},
	},
	{
		id: "last_contact_at",
		header: "Last contact",
		accessorKey: "last_contact_at",
		cell: ({ getValue }) => {
			const v = getValue<string | null>()
			return v ? (
				<span className="text-sm font-base">{new Date(v).toLocaleDateString()}</span>
			) : (
				<span className="text-xs font-base text-foreground/40">—</span>
			)
		},
	},
	{
		id: "actions",
		header: "",
		size: 80,
		cell: ({ row }) => (
			<Button variant="ghost" size="sm" asChild>
				<Link to="/people/$personId/edit" params={{ personId: String(row.original.id) }}>
					Edit
				</Link>
			</Button>
		),
	},
]

export function PeopleTable({ q = "", labels = [], page = 1, page_size = 20, onSearchChange }: PeopleTableProps) {
	const [localQ, setLocalQ] = useState(q)
	const debouncedQ = useDebounce(localQ, 300)
	const isFirst = useRef(true)

	useEffect(() => {
		if (isFirst.current) { isFirst.current = false; return }
		onSearchChange(debouncedQ)
	}, [debouncedQ, onSearchChange])

	useEffect(() => { setLocalQ(q) }, [q])

	const { data, isLoading } = useQuery({
		queryKey: keys.people.list({
			q: debouncedQ || undefined,
			labels: labels.length ? labels : undefined,
			page,
			page_size,
		}),
		queryFn: () =>
			listPeople({
				q: debouncedQ || undefined,
				labels: labels.length ? labels : undefined,
				page,
				page_size,
			}),
	})

	// Prefetch labels for filter chips (consumed by parent page if needed)
	useQuery({ queryKey: keys.labels.list(), queryFn: listLabels })

	const rows = data?.items ?? []

	return (
		<div className="space-y-3">
			<Input
				placeholder="Search people…"
				value={localQ}
				onChange={(e) => setLocalQ(e.target.value)}
				className="max-w-xs"
			/>
			<DataTable
				columns={columns}
				data={rows}
				pageSize={page_size}
				emptyState={
					isLoading ? (
						<span className="text-sm font-base text-foreground/50">Loading…</span>
					) : (
						<span className="text-sm font-base text-foreground/50">No people found.</span>
					)
				}
			/>
		</div>
	)
}
