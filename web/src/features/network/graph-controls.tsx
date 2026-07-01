import type { NetworkOnlyMineDepth } from "#/lib/format-datetime";
import type { ColorBy } from "./graph-types";

interface GraphControlsProps {
	title?: string;
	colorBy: ColorBy;
	onColorByChange: (v: ColorBy) => void;
	groups: string[];
	activeGroup: string | null;
	onGroupChange: (g: string | null) => void;
	relTypes: string[];
	activeRelType: string | null;
	onRelTypeChange: (t: string | null) => void;
	showAvatar: boolean;
	onShowAvatarChange: (v: boolean) => void;
	showOnlyMine?: boolean;
	onShowOnlyMineChange?: (v: boolean) => void;
	onlyMineDepth?: NetworkOnlyMineDepth;
	onOnlyMineDepthChange?: (v: NetworkOnlyMineDepth) => void;
	showUnconnected?: boolean;
	onShowUnconnectedChange?: (v: boolean) => void;
	onRecenter: () => void;
}

export function GraphControls({
	title,
	colorBy,
	onColorByChange,
	groups,
	activeGroup,
	onGroupChange,
	relTypes,
	activeRelType,
	onRelTypeChange,
	showAvatar,
	onShowAvatarChange,
	showOnlyMine,
	onShowOnlyMineChange,
	onlyMineDepth,
	onOnlyMineDepthChange,
	showUnconnected,
	onShowUnconnectedChange,
	onRecenter,
}: GraphControlsProps) {
	return (
		<div className="flex flex-wrap items-center gap-x-3 gap-y-1.5 border-b border-zinc-200 px-3 py-2.5">
			{title && (
				<span className="mr-1 text-[14px] font-semibold text-zinc-900">
					{title}
				</span>
			)}

			{/* Group filter */}
			{groups.length > 0 && (
				<div className="flex items-center gap-1.5 text-[11px] text-zinc-500">
					<span className="font-medium">Group</span>
					<select
						value={activeGroup ?? ""}
						onChange={(e) => onGroupChange(e.target.value || null)}
						className="h-7 rounded border border-zinc-200 bg-white px-1.5 text-[11px] text-zinc-700 focus:outline-none focus:ring-1 focus:ring-indigo-500"
					>
						<option value="">Any</option>
						{groups.map((g) => (
							<option key={g} value={g}>
								{g}
							</option>
						))}
					</select>
				</div>
			)}

			{/* Rel type filter */}
			{relTypes.length > 0 && (
				<div className="flex items-center gap-1.5 text-[11px] text-zinc-500">
					<span className="font-medium">Type</span>
					<select
						value={activeRelType ?? ""}
						onChange={(e) => onRelTypeChange(e.target.value || null)}
						className="h-7 rounded border border-zinc-200 bg-white px-1.5 text-[11px] text-zinc-700 focus:outline-none focus:ring-1 focus:ring-indigo-500"
					>
						<option value="">Any</option>
						{relTypes.map((t) => (
							<option key={t} value={t}>
								{t}
							</option>
						))}
					</select>
				</div>
			)}

			{/* Color by segmented control */}
			<div className="flex items-center gap-1.5 text-[11px] text-zinc-500">
				<span className="font-medium">Color by</span>
				<div className="flex items-center gap-0.5 rounded border border-zinc-200 bg-white p-0.5">
					{(["labels", "type"] as ColorBy[]).map((v) => (
						<button
							key={v}
							type="button"
							onClick={() => onColorByChange(v)}
							className={`min-h-0 rounded px-2 py-0.5 text-[11px] font-medium transition-colors ${
								colorBy === v
									? "bg-indigo-600 text-white"
									: "text-zinc-600 hover:bg-zinc-100"
							}`}
						>
							{v === "labels" ? "Labels" : "Rel. type"}
						</button>
					))}
				</div>
			</div>

			{/* Show avatars checkbox */}
			<label className="flex cursor-pointer items-center gap-1.5 text-[11px] text-zinc-600 select-none">
				<input
					type="checkbox"
					checked={showAvatar}
					onChange={(e) => onShowAvatarChange(e.target.checked)}
					className="h-3.5 w-3.5 rounded border-zinc-300 accent-indigo-600"
				/>
				Avatars
			</label>

			{/* Only my connections (optional — network page only) */}
			{onShowOnlyMineChange && (
				<div className="flex items-center gap-1.5">
					<label className="flex cursor-pointer items-center gap-1.5 text-[11px] text-zinc-600 select-none">
						<input
							type="checkbox"
							checked={showOnlyMine ?? false}
							onChange={(e) => onShowOnlyMineChange(e.target.checked)}
							className="h-3.5 w-3.5 rounded border-zinc-300 accent-indigo-600"
						/>
						Only mine
					</label>
					{showOnlyMine && onOnlyMineDepthChange && (
						<div className="flex items-center gap-1.5">
							<div className="flex items-center gap-0.5 rounded border border-zinc-200 bg-white p-0.5">
								{(["direct", "alter"] as NetworkOnlyMineDepth[]).map((v) => (
									<button
										key={v}
										type="button"
										onClick={() => onOnlyMineDepthChange(v)}
										className={`min-h-0 rounded px-2 py-0.5 text-[11px] font-medium transition-colors ${
											(onlyMineDepth ?? "direct") === v
												? "bg-indigo-600 text-white"
												: "text-zinc-600 hover:bg-zinc-100"
										}`}
									>
										{v === "direct" ? "Direct" : "Alters"}
									</button>
								))}
							</div>
							{(onlyMineDepth ?? "direct") === "alter" && (
								<span className="text-[10px] text-zinc-400 italic">
									indirect connections
								</span>
							)}
						</div>
					)}
				</div>
			)}

			{/* Show unconnected people (optional — network page only) */}
			{onShowUnconnectedChange && (
				<label className="flex cursor-pointer items-center gap-1.5 text-[11px] text-zinc-600 select-none">
					<input
						type="checkbox"
						checked={showUnconnected ?? true}
						onChange={(e) => onShowUnconnectedChange(e.target.checked)}
						className="h-3.5 w-3.5 rounded border-zinc-300 accent-indigo-600"
					/>
					Unconnected
				</label>
			)}

			{/* Recenter — pushed to far right */}
			<button
				type="button"
				onClick={onRecenter}
				title="Recenter"
				aria-label="Recenter graph"
				className="ml-auto flex h-7 w-7 items-center justify-center rounded border border-zinc-200 bg-white text-zinc-500 hover:border-indigo-400 hover:text-indigo-600"
			>
				<svg
					aria-hidden="true"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					strokeWidth={2}
					strokeLinecap="round"
					strokeLinejoin="round"
					className="h-3.5 w-3.5"
				>
					<circle cx="12" cy="12" r="3" />
					<path d="M12 2v3M12 19v3M2 12h3M19 12h3" />
				</svg>
			</button>
		</div>
	);
}
