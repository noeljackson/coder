import { useQuery } from "@tanstack/react-query";
import { API } from "api/api";
import { useDeploymentConfig } from "modules/management/DeploymentConfigProvider";
import type { FC } from "react";
import { pageTitle } from "utils/page";
import { ExternalAuthSettingsPageView } from "./ExternalAuthSettingsPageView";

const ExternalAuthSettingsPage: FC = () => {
	const { deploymentConfig } = useDeploymentConfig();

	const { data: dynamicProviders, refetch } = useQuery({
		queryKey: ["externalAuthProviders"],
		queryFn: () => API.getExternalAuthProviders(),
	});

	return (
		<>
			<title>{pageTitle("External Authentication Settings")}</title>

			<ExternalAuthSettingsPageView
				config={deploymentConfig.config}
				dynamicProviders={dynamicProviders}
				onProviderDeleted={() => refetch()}
			/>
		</>
	);
};

export default ExternalAuthSettingsPage;
