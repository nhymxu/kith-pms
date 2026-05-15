import { Plus, Trash2 } from "lucide-react"
import { Button } from "#/components/ui/button"
import { Input } from "#/components/ui/input"
import { Label } from "#/components/ui/label"
import type { Location } from "#/schemas/person"

type LocationRow = Omit<Location, "id" | "person_id">

interface Props {
	value: LocationRow[]
	onChange: (rows: LocationRow[]) => void
}

const LOCATION_TYPES = ["home", "work", "other"]

export function PersonFormLocations({ value, onChange }: Props) {
	function add() {
		onChange([
			...value,
			{ type: "home", address: "", city: "", country: "", postal_code: "", position: value.length },
		])
	}

	function remove(i: number) {
		onChange(value.filter((_, idx) => idx !== i))
	}

	function update(i: number, field: keyof LocationRow, v: string | number) {
		onChange(value.map((row, idx) => (idx === i ? { ...row, [field]: v } : row)))
	}

	return (
		<div className="space-y-3">
			<div className="flex items-center justify-between">
				<Label className="text-sm font-heading">Locations</Label>
				<Button type="button" variant="neutral" size="sm" onClick={add}>
					<Plus className="size-3" /> Add
				</Button>
			</div>
			{value.length === 0 && (
				<p className="text-xs font-base text-foreground/50">No locations yet.</p>
			)}
			{value.map((row, i) => (
				<div key={i} className="border-2 border-border rounded-base p-3 space-y-2">
					<div className="flex items-center justify-between">
						<select
							className="h-9 border-2 border-border rounded-base bg-background px-2 text-sm font-base"
							value={row.type}
							onChange={(e) => update(i, "type", e.target.value)}
						>
							{LOCATION_TYPES.map((t) => (
								<option key={t} value={t}>{t}</option>
							))}
						</select>
						<Button
							type="button"
							variant="destructive"
							size="icon"
							onClick={() => remove(i)}
						>
							<Trash2 className="size-3" />
						</Button>
					</div>
					<Input
						placeholder="Address"
						value={row.address}
						onChange={(e) => update(i, "address", e.target.value)}
					/>
					<div className="grid grid-cols-3 gap-2">
						<Input
							placeholder="City"
							value={row.city}
							onChange={(e) => update(i, "city", e.target.value)}
						/>
						<Input
							placeholder="Country"
							value={row.country}
							onChange={(e) => update(i, "country", e.target.value)}
						/>
						<Input
							placeholder="Postal code"
							value={row.postal_code}
							onChange={(e) => update(i, "postal_code", e.target.value)}
						/>
					</div>
				</div>
			))}
		</div>
	)
}
