import type { Table } from "@tanstack/react-table";
import { Search, X } from "lucide-react";
import { Button } from "#/components/ui/button";
import { Input } from "#/components/ui/input";

interface DataTableToolbarProps<T> {
	table: Table<T>;
	globalFilter: string;
	onGlobalFilterChange: (value: string) => void;
	children?: React.ReactNode;
}

export function DataTableToolbar<T>({
	table: _table,
	globalFilter,
	onGlobalFilterChange,
	children,
}: DataTableToolbarProps<T>) {
	const hasFilter = globalFilter.length > 0;

	return (
		<div className="flex items-center justify-between gap-3 py-3">
			<div className="relative flex-1 max-w-sm">
				<Search className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-foreground/50" />
				<Input
					placeholder="Search…"
					value={globalFilter}
					onChange={(e) => onGlobalFilterChange(e.target.value)}
					className="pl-9"
				/>
				{hasFilter && (
					<Button
						variant="ghost"
						size="icon"
						className="absolute right-1 top-1/2 -translate-y-1/2 size-7"
						onClick={() => onGlobalFilterChange("")}
						aria-label="Clear search"
					>
						<X className="size-3" />
					</Button>
				)}
			</div>
			{children && <div className="flex items-center gap-2">{children}</div>}
		</div>
	);
}
