import { Link } from "@tanstack/react-router";
import { formatBirthdayLabel, formatRelativeDate } from "./graph-date-format";
import type { GraphNode } from "./graph-types";

interface GraphNodeProfileCardProps {
	node: GraphNode;
	/** Canvas-container-relative pixel position of the clicked node centre. */
	posX: number;
	posY: number;
	showAvatar: boolean;
	groupColorMap: Map<string, string>;
	onClose: () => void;
}

function NodeAvatarDisc({
	node,
	showAvatar,
	groupColorMap,
	size,
}: {
	node: GraphNode;
	showAvatar: boolean;
	groupColorMap: Map<string, string>;
	size: number;
}) {
	const color = node.group
		? (groupColorMap.get(node.group) ?? "#71717a")
		: "#71717a";
	const initial = (node.name?.[0] ?? "?").toUpperCase();

	if (showAvatar && node.avatar) {
		return (
			<img
				src={node.avatar}
				alt={node.name}
				width={size}
				height={size}
				className="flex-none rounded-full object-cover"
				style={{ width: size, height: size }}
				onError={(e) => {
					(e.currentTarget as HTMLImageElement).style.display = "none";
				}}
			/>
		);
	}

	return (
		<div
			className="flex flex-none items-center justify-center rounded-full text-white font-semibold"
			style={{
				width: size,
				height: size,
				background: color,
				fontSize: size * 0.4,
			}}
		>
			{initial}
		</div>
	);
}

export function GraphNodeProfileCard({
	node,
	posX,
	posY,
	showAvatar,
	groupColorMap,
	onClose,
}: GraphNodeProfileCardProps) {
	const birthday = node.date_of_birth
		? formatBirthdayLabel(node.date_of_birth)
		: null;
	const lastContacted = node.last_contact_at
		? formatRelativeDate(node.last_contact_at)
		: null;

	return (
		<div
			className="absolute z-20 w-[220px] overflow-hidden rounded-md border border-zinc-200 bg-white shadow-lg"
			style={{
				left: posX,
				top: posY,
				transform: "translate(-50%, -100%) translateY(-12px)",
			}}
			onPointerDown={(e) => e.stopPropagation()}
		>
			{/* Header */}
			<div className="flex items-start gap-2.5 p-3">
				<NodeAvatarDisc
					node={node}
					showAvatar={showAvatar}
					groupColorMap={groupColorMap}
					size={36}
				/>
				<div className="min-w-0 flex-1">
					<div className="truncate text-[13px] font-semibold text-zinc-900">
						{node.name}
					</div>
					{node.nickname && (
						<div className="truncate text-[11px] text-zinc-400 italic">
							{node.nickname}
						</div>
					)}
					{node.groups && node.groups.length > 0 && (
						<div className="flex flex-wrap gap-1 mt-0.5">
							{node.groups.map((g) => (
								<span
									key={g}
									className="rounded bg-zinc-100 px-1 py-0.5 text-[10px] font-medium text-zinc-500"
								>
									{g}
								</span>
							))}
						</div>
					)}
				</div>
				<button
					type="button"
					onClick={onClose}
					className="flex-none text-[16px] leading-none text-zinc-400 hover:text-zinc-600"
					aria-label="Close"
				>
					×
				</button>
			</div>

			{/* Details */}
			{(birthday || lastContacted) && (
				<div className="space-y-1 border-t border-zinc-100 px-3 py-2">
					{birthday && (
						<div className="flex items-center gap-1.5 text-[11px] text-zinc-500">
							<span>🎂</span>
							<span>{birthday}</span>
						</div>
					)}
					{lastContacted && (
						<div className="flex items-center gap-1.5 text-[11px] text-zinc-500">
							<span>🕐</span>
							<span>Last contacted {lastContacted}</span>
						</div>
					)}
				</div>
			)}

			{/* Action */}
			<div className="border-t border-zinc-100 px-3 py-2.5">
				<Link
					to="/people/$personId"
					params={{ personId: String(node.id) }}
					target="_blank"
					rel="noopener noreferrer"
					className="text-[12px] font-medium text-indigo-600 hover:underline"
				>
					Open profile →
				</Link>
			</div>
		</div>
	);
}
