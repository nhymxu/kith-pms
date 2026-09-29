import { Button } from "#/components/ui/button";

const inputClass =
	"h-9 border-field-bw border-field-line rounded-md bg-field px-3 text-[13px] focus:outline-none focus:ring-2 focus:ring-ring";

interface NotesDateRangeProps {
	idPrefix: string;
	from: string;
	to: string;
	onChange: (from: string, to: string) => void;
}

export function NotesDateRange({
	idPrefix,
	from,
	to,
	onChange,
}: NotesDateRangeProps) {
	return (
		<div className="flex flex-wrap gap-3 items-end">
			<div className="space-y-1">
				<label
					htmlFor={`${idPrefix}-from`}
					className="text-[11px] font-medium text-sub"
				>
					From
				</label>
				<input
					id={`${idPrefix}-from`}
					type="date"
					value={from}
					onChange={(e) => onChange(e.target.value, to)}
					className={inputClass}
				/>
			</div>
			<div className="space-y-1">
				<label
					htmlFor={`${idPrefix}-to`}
					className="text-[11px] font-medium text-sub"
				>
					To
				</label>
				<input
					id={`${idPrefix}-to`}
					type="date"
					value={to}
					onChange={(e) => onChange(from, e.target.value)}
					className={inputClass}
				/>
			</div>
			{(from || to) && (
				<Button variant="neutral" size="sm" onClick={() => onChange("", "")}>
					Clear dates
				</Button>
			)}
		</div>
	);
}
