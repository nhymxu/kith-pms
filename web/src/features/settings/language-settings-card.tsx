import { useMutation, type useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "#/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "#/components/ui/card";
import { Label } from "#/components/ui/label";
import { updateSettings } from "#/endpoints/settings";
import {
	getUserPrefs,
	type Language,
	saveUserPrefs,
} from "#/lib/format-datetime";
import type { UserSettings } from "#/schemas/settings";

// Native names, never translated, so a user can find their language.
const LANGUAGE_OPTIONS: { value: Language; label: string }[] = [
	{ value: "en", label: "English" },
	{ value: "vi", label: "Tiếng Việt" },
];

export function LanguageSettingsCard({
	apiSettings,
	isPlaceholderData,
	buildPayload,
	queryClient,
}: {
	apiSettings: UserSettings | undefined;
	isPlaceholderData: boolean;
	buildPayload: (
		overrides?: Partial<Parameters<typeof updateSettings>[0]>,
	) => Parameters<typeof updateSettings>[0];
	queryClient: ReturnType<typeof useQueryClient>;
}) {
	const { t } = useTranslation("settings");
	const [language, setLanguage] = useState<Language>(
		() => getUserPrefs().language,
	);

	const [synced, setSynced] = useState(false);
	if (apiSettings && !isPlaceholderData && !synced) {
		setLanguage(apiSettings.language);
		setSynced(true);
	}

	const mutation = useMutation({
		mutationFn: () => updateSettings(buildPayload({ language })),
		onSuccess: (updated) => {
			saveUserPrefs({ language: updated.language as Language });
			queryClient.setQueryData(["settings"], updated);
		},
	});

	return (
		<Card>
			<CardHeader>
				<CardTitle className="text-[14px] font-medium text-ink">
					{t("language.title")}
				</CardTitle>
				<CardDescription className="text-[12px] text-sub">
					{t("language.description")}
				</CardDescription>
			</CardHeader>
			<CardContent className="space-y-4">
				<div className="space-y-2">
					<Label className="text-[13px]">{t("language.label")}</Label>
					<div className="space-y-1.5">
						{LANGUAGE_OPTIONS.map((opt) => (
							<label
								key={opt.value}
								className="flex items-center gap-3 cursor-pointer"
							>
								<input
									type="radio"
									name="language"
									value={opt.value}
									checked={language === opt.value}
									onChange={() => setLanguage(opt.value)}
									className="accent-accent"
								/>
								<span className="text-[13px] text-ink">{opt.label}</span>
							</label>
						))}
					</div>
				</div>

				<Button
					onClick={() => mutation.mutate()}
					size="sm"
					disabled={isPlaceholderData || mutation.isPending}
				>
					{mutation.isPending
						? t("language.saving")
						: mutation.isSuccess
							? t("language.saved")
							: t("language.save")}
				</Button>
				{mutation.isError && (
					<p className="text-[12px] text-danger-fg">{t("language.error")}</p>
				)}
			</CardContent>
		</Card>
	);
}
