const PAGE_SIZE_OPTIONS = [10, 25, 50, 100, 200] as const;

interface PageSizeSelectorProps {
	value: number;
	hasOverride: boolean;
	onChange: (n: number) => void;
	onReset: () => void;
}

export function PageSizeSelector({
	value,
	hasOverride,
	onChange,
	onReset,
}: PageSizeSelectorProps) {
	return (
		<div className="flex items-center gap-1.5">
			<label htmlFor="page-size-select" className="text-[12px] text-zinc-500">
				Rows
			</label>
			<select
				id="page-size-select"
				value={value}
				onChange={(e) => onChange(Number(e.target.value))}
				className="h-8 border border-zinc-200 rounded-md bg-white px-2 text-[12px] focus:outline-none focus:ring-2 focus:ring-indigo-600"
			>
				{PAGE_SIZE_OPTIONS.map((n) => (
					<option key={n} value={n}>
						{n}
					</option>
				))}
			</select>
			{hasOverride && (
				<button
					type="button"
					onClick={onReset}
					className="text-[11px] text-zinc-400 hover:text-zinc-700"
				>
					Reset
				</button>
			)}
		</div>
	);
}
