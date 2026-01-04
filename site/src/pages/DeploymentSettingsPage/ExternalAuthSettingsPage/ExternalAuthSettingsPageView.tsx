import { css } from "@emotion/react";
import type {
	DeploymentValues,
	ExternalAuthConfig,
	ExternalAuthProviderConfig,
} from "api/typesGenerated";
import { Alert } from "components/Alert/Alert";
import { PremiumBadge } from "components/Badges/Badges";
import {
	SettingsHeader,
	SettingsHeaderDescription,
	SettingsHeaderDocsLink,
	SettingsHeaderTitle,
} from "components/SettingsHeader/SettingsHeader";
import {
	Table,
	TableBody,
	TableCell,
	TableHead,
	TableHeader,
	TableRow,
} from "components/Table/Table";
import type { FC } from "react";
import { docs } from "utils/docs";
import { CreateGitHubAppDialog } from "./CreateGitHubAppDialog";

type ExternalAuthSettingsPageViewProps = {
	config: DeploymentValues;
	dynamicProviders?: ExternalAuthProviderConfig[];
	onProviderDeleted?: () => void;
};

export const ExternalAuthSettingsPageView: FC<
	ExternalAuthSettingsPageViewProps
> = ({ config, dynamicProviders = [], onProviderDeleted }) => {
	return (
		<>
			<SettingsHeader
				actions={
					<div className="flex items-center gap-2">
						<CreateGitHubAppDialog onSuccess={onProviderDeleted} />
						<SettingsHeaderDocsLink href={docs("/admin/external-auth")} />
					</div>
				}
			>
				<SettingsHeaderTitle>External Authentication</SettingsHeaderTitle>
				<SettingsHeaderDescription>
					Coder integrates with GitHub, GitLab, BitBucket, Azure Repos, and
					OpenID Connect to authenticate developers with external services.
				</SettingsHeaderDescription>
			</SettingsHeader>

			<video
				autoPlay
				muted
				loop
				playsInline
				src="/external-auth.mp4"
				style={{
					maxWidth: "100%",
					borderRadius: 4,
				}}
			/>

			<div
				css={{
					marginTop: 24,
					marginBottom: 24,
				}}
			>
				<Alert severity="info" actions={<PremiumBadge key="enterprise" />}>
					Integrating with multiple External authentication providers is an
					Premium feature.
				</Alert>
			</div>

			<Table
				css={css`
            & td {
              padding-top: 24px;
              padding-bottom: 24px;
            }

            & td:last-child,
            & th:last-child {
              padding-left: 32px;
            }
          `}
			>
				<TableHeader>
					<TableRow>
						<TableHead className="w-1/3">ID</TableHead>
						<TableHead className="w-1/3">Client ID</TableHead>
						<TableHead className="w-1/3">Match</TableHead>
					</TableRow>
				</TableHeader>
				<TableBody>
					{((config.external_auth === null ||
						config.external_auth?.length === 0) && (
						<TableRow>
							<TableCell colSpan={999}>
								<div css={{ textAlign: "center" }}>
									No providers have been configured!
								</div>
							</TableCell>
						</TableRow>
					)) ||
						config.external_auth?.map((git: ExternalAuthConfig) => {
							const name = git.id || git.type;
							return (
								<TableRow key={name}>
									<TableCell>{name}</TableCell>
									<TableCell>{git.client_id}</TableCell>
									<TableCell>{git.regex || "Not Set"}</TableCell>
								</TableRow>
							);
						})}
				</TableBody>
			</Table>

			{dynamicProviders.length > 0 && (
				<>
					<h3 className="mb-4 mt-8 text-lg font-semibold">
						Dynamically Configured Providers
					</h3>
					<p className="mb-4 text-sm text-content-secondary">
						These providers were created through the UI and are stored in the
						database.
					</p>
					<Table
						css={css`
							& td {
								padding-top: 24px;
								padding-bottom: 24px;
							}

							& td:last-child,
							& th:last-child {
								padding-left: 32px;
							}
						`}
					>
						<TableHeader>
							<TableRow>
								<TableHead className="w-1/4">ID</TableHead>
								<TableHead className="w-1/4">Type</TableHead>
								<TableHead className="w-1/4">Client ID</TableHead>
								<TableHead className="w-1/4">Created</TableHead>
							</TableRow>
						</TableHeader>
						<TableBody>
							{dynamicProviders.map((provider) => (
								<TableRow key={provider.id}>
									<TableCell>
										<div className="flex items-center gap-2">
											{provider.display_icon && (
												<img
													src={provider.display_icon}
													alt=""
													className="h-5 w-5"
												/>
											)}
											{provider.display_name || provider.id}
										</div>
									</TableCell>
									<TableCell>{provider.type}</TableCell>
									<TableCell>{provider.client_id}</TableCell>
									<TableCell>
										{new Date(provider.created_at).toLocaleDateString()}
									</TableCell>
								</TableRow>
							))}
						</TableBody>
					</Table>
				</>
			)}
		</>
	);
};
