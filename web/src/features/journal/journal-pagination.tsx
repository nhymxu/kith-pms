import { ChevronLeft, ChevronRight } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import {
	Popover,
	PopoverContent,
	PopoverTrigger,
} from "#/components/ui/popover";

interface JournalPaginationProps {
	page: number;
	pageSize: number;
	total: number;
	onPageChange: (page: number) => void;
}

export function JournalPagination({
	page,
	pageSize,
	total,
	onPageChange,
}: JournalPaginationProps) {
	const pageCount = Math.max(1, Math.ceil(total / pageSize));

	return (
		<div className="flex justify-center pt-2">
			<div className="inline-flex items-stretch rounded-md border border-zinc-200 overflow-hidden">
				<button
					type="button"
					disabled={page <= 1}
					onClick={() => onPageChange(page - 1)}
					className="flex items-center gap-1.5 px-3 h-9 text-[13px] font-medium text-zinc-700 hover:bg-zinc-50 disabled:opacity-40 disabled:pointer-events-none border-r border-zinc-200"
				>
					<ChevronLeft className="size-4" /> Prev
				</button>

				<JumpPopover
					page={page}
					pageCount={pageCount}
					onPageChange={onPageChange}
				/>

				<button
					type="button"
					disabled={page >= pageCount}
					onClick={() => onPageChange(page + 1)}
					className="flex items-center gap-1.5 px-3 h-9 text-[13px] font-medium text-zinc-700 hover:bg-zinc-50 disabled:opacity-40 disabled:pointer-events-none border-l border-zinc-200"
				>
					Next <ChevronRight className="size-4" />
				</button>
			</div>
		</div>
	);
}

interface JumpPopoverProps {
	page: number;
	pageCount: number;
	onPageChange: (page: number) => void;
}

function JumpPopover({ page, pageCount, onPageChange }: JumpPopoverProps) {
	const [open, setOpen] = useState(false);
	const [value, setValue] = useState(String(page));
	const inputRef = useRef<HTMLInputElement>(null);

	useEffect(() => {
		if (open) inputRef.current?.focus();
	}, [open]);

	const submit = () => {
		const parsed = Number.parseInt(value, 10);
		if (!Number.isNaN(parsed)) {
			onPageChange(Math.min(Math.max(parsed, 1), pageCount));
		}
		setOpen(false);
	};

	return (
		<Popover
			open={open}
			onOpenChange={(next) => {
				setOpen(next);
				if (next) setValue(String(page));
			}}
		>
			<PopoverTrigger className="flex items-center justify-center min-w-16 px-3 h-9 text-[13px] font-semibold bg-indigo-600 text-white hover:bg-indigo-700">
				{page}/{pageCount}
			</PopoverTrigger>
			<PopoverContent className="w-44">
				<label
					htmlFor="journal-page-jump"
					className="text-[11px] font-medium text-zinc-500"
				>
					Jump to page
				</label>
				<div className="flex items-center gap-2 mt-1.5">
					<input
						id="journal-page-jump"
						type="number"
						min={1}
						max={pageCount}
						ref={inputRef}
						value={value}
						onChange={(e) => setValue(e.target.value)}
						onKeyDown={(e) => {
							if (e.key === "Enter") {
								e.preventDefault();
								submit();
							}
						}}
						className="h-8 w-full border border-zinc-200 rounded-md bg-white px-2 text-[13px] focus:outline-none focus:ring-2 focus:ring-indigo-600"
					/>
					<button
						type="button"
						onClick={submit}
						className="h-8 px-3 rounded-md bg-indigo-600 text-white text-[13px] font-medium hover:bg-indigo-700 shrink-0"
					>
						Go
					</button>
				</div>
			</PopoverContent>
		</Popover>
	);
}
