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
			<div className="inline-flex items-stretch rounded-md border-bw border-line overflow-hidden">
				<button
					type="button"
					disabled={page <= 1}
					onClick={() => onPageChange(page - 1)}
					className="flex items-center gap-1.5 px-3 h-9 text-[13px] font-medium text-ink hover:bg-chip disabled:opacity-40 disabled:pointer-events-none border-r border-line"
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
					className="flex items-center gap-1.5 px-3 h-9 text-[13px] font-medium text-ink hover:bg-chip disabled:opacity-40 disabled:pointer-events-none border-l border-line"
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
			<PopoverTrigger className="flex items-center justify-center min-w-16 px-3 h-9 text-[13px] font-semibold bg-accent text-accent-foreground hover:bg-accent/90">
				{page}/{pageCount}
			</PopoverTrigger>
			<PopoverContent className="w-44">
				<label
					htmlFor="journal-page-jump"
					className="text-[11px] font-medium text-sub"
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
						className="h-8 w-full border-field-bw border-field-line rounded-md bg-field px-2 text-[13px] focus:outline-none focus:ring-2 focus:ring-ring"
					/>
					<button
						type="button"
						onClick={submit}
						className="h-8 px-3 rounded-md bg-accent text-accent-foreground text-[13px] font-medium hover:bg-accent/90 shrink-0"
					>
						Go
					</button>
				</div>
			</PopoverContent>
		</Popover>
	);
}
