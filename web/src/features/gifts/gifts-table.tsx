// Gifts table: columns for image, title, person, date, amount, debt direction badge

import { Link } from "@tanstack/react-router";
import { useMemo } from "react";
import {
	sortableHeader,
	type TableColumn,
	valueCell,
} from "#/components/data-table/column-helpers";
import { DataTable } from "#/components/data-table/data-table";
import { Pill } from "#/components/ui/pill";
import { formatDate } from "#/lib/format-datetime";
import type { GiftWithPerson } from "#/schemas/gift";

interface GiftsTableProps {
	data: GiftWithPerson[];
	toolbarActions?: React.ReactNode;
	pageSizeSelector?: React.ReactNode;
	pageSize?: number;
	totalCount?: number;
	pageIndex?: number;
	onPageChange?: (pageIndex: number) => void;
}

function DebtBadge({
	debtType,
	direction,
}: {
	debtType: string;
	direction: string;
}) {
	if (direction === "given") return <Pill variant="plain">Given</Pill>;
	if (direction === "received") return <Pill variant="accent">Received</Pill>;
	if (debtType === "i_owe") return <Pill variant="warning">I owe</Pill>;
	if (debtType === "they_owe") return <Pill variant="success">They owe</Pill>;
	return <Pill variant="plain">Planned</Pill>;
}

export function GiftsTable({
	data,
	toolbarActions,
	pageSizeSelector,
	pageSize,
	totalCount,
	pageIndex,
	onPageChange,
}: GiftsTableProps) {
	const columns = useMemo<TableColumn<GiftWithPerson>[]>(
		() => [
			{
				id: "image",
				header: "",
				size: 48,
				cell: ({ row }) => {
					const gift = row.original;
					if (!gift.image_path) {
						return (
							<div className="w-8 h-8 rounded-base bg-chip border-bw border-line" />
						);
					}
					return (
						<img
							src={`/v1/gifts/${gift.id}/image`}
							alt=""
							className="w-8 h-8 rounded-base object-cover border-bw border-line"
						/>
					);
				},
			},
			{
				id: "title",
				accessorKey: "title",
				header: sortableHeader<GiftWithPerson>("Gift"),
				enableSorting: true,
				cell: valueCell<GiftWithPerson, string>((val, row) => (
					<Link
						to="/gifts/$giftId"
						params={{ giftId: String(row.id) }}
						className="text-[13px] text-ink hover:text-accent-text hover:underline"
					>
						{val}
					</Link>
				)),
			},
			{
				id: "person_name",
				accessorKey: "person_name",
				header: sortableHeader<GiftWithPerson>("Person"),
				enableSorting: true,
				cell: valueCell<GiftWithPerson, string>((val, row) =>
					row.person_id ? (
						<Link
							to="/people/$personId"
							params={{ personId: String(row.person_id) }}
							className="text-accent-text hover:underline"
						>
							{val}
						</Link>
					) : (
						<span>{val}</span>
					),
				),
			},
			{
				id: "date",
				accessorKey: "date",
				header: sortableHeader<GiftWithPerson>("Date"),
				enableSorting: true,
				cell: valueCell<GiftWithPerson, string>((val) =>
					val ? (
						<span className="font-mono text-[12px] text-sub">
							{formatDate(val)}
						</span>
					) : (
						<span className="text-sub/60">—</span>
					),
				),
			},
			{
				id: "amount",
				accessorKey: "amount_cents",
				header: "Amount",
				cell: valueCell<GiftWithPerson, number | null>((val, row) =>
					val != null ? (
						<span className="font-mono text-[12px] text-ink">
							{row.currency || "USD"} {(val / 100).toFixed(2)}
						</span>
					) : (
						<span className="text-sub/60">—</span>
					),
				),
			},
			{
				id: "debt",
				header: "Direction",
				cell: ({ row }) => (
					<DebtBadge
						debtType={row.original.debt_type ?? ""}
						direction={row.original.direction}
					/>
				),
			},
		],
		[],
	);

	return (
		<DataTable
			columns={columns}
			data={data}
			toolbarActions={toolbarActions}
			pageSizeSelector={pageSizeSelector}
			pageSize={pageSize}
			totalCount={totalCount}
			pageIndex={pageIndex}
			onPageChange={onPageChange}
			emptyState={
				<span className="text-sm text-foreground/50">No gifts yet.</span>
			}
		/>
	);
}
