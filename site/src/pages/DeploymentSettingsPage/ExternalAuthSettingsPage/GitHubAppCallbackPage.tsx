import { useMutation } from "@tanstack/react-query";
import { CheckCircle, Github, XCircle } from "lucide-react";
import { useEffect, useState } from "react";
import { useNavigate, useSearchParams } from "react-router";
import { API } from "api/api";
import type { ExternalAuthProviderConfig } from "api/typesGenerated";
import { Alert } from "components/Alert/Alert";
import { Spinner } from "components/Spinner/Spinner";

export const GitHubAppCallbackPage: React.FC = () => {
	const [searchParams] = useSearchParams();
	const navigate = useNavigate();
	const [provider, setProvider] = useState<ExternalAuthProviderConfig | null>(
		null,
	);

	const code = searchParams.get("code");
	const state = searchParams.get("state");

	const completeManifest = useMutation({
		mutationFn: async () => {
			if (!code || !state) {
				throw new Error("Missing code or state parameter");
			}
			return API.completeGitHubAppManifest({ code, state });
		},
		onSuccess: (data) => {
			setProvider(data);
			// Redirect to external auth settings after a delay.
			setTimeout(() => {
				navigate("/deployment/external-auth");
			}, 3000);
		},
	});

	useEffect(() => {
		if (code && state && !completeManifest.isPending && !provider) {
			completeManifest.mutate();
		}
	}, [code, state, completeManifest, provider]);

	if (!code || !state) {
		return (
			<div className="flex min-h-screen items-center justify-center">
				<div className="max-w-md text-center">
					<div className="mb-4 flex justify-center">
						<XCircle className="h-16 w-16 text-content-danger" />
					</div>
					<h1 className="mb-2 text-2xl font-semibold">Invalid Callback</h1>
					<p className="text-content-secondary">
						Missing required parameters. Please try creating the GitHub App
						again.
					</p>
				</div>
			</div>
		);
	}

	if (completeManifest.isPending) {
		return (
			<div className="flex min-h-screen items-center justify-center">
				<div className="max-w-md text-center">
					<div className="mb-4 flex justify-center">
						<Github className="h-16 w-16 text-content-primary" />
					</div>
					<h1 className="mb-2 text-2xl font-semibold">
						Completing GitHub App Setup
					</h1>
					<p className="mb-4 text-content-secondary">
						Please wait while we finish configuring your GitHub App...
					</p>
					<Spinner />
				</div>
			</div>
		);
	}

	if (completeManifest.isError) {
		return (
			<div className="flex min-h-screen items-center justify-center">
				<div className="max-w-md">
					<div className="mb-4 flex justify-center">
						<XCircle className="h-16 w-16 text-content-danger" />
					</div>
					<h1 className="mb-2 text-center text-2xl font-semibold">
						Setup Failed
					</h1>
					<Alert severity="error">
						{completeManifest.error instanceof Error
							? completeManifest.error.message
							: "An unexpected error occurred"}
					</Alert>
				</div>
			</div>
		);
	}

	if (provider) {
		return (
			<div className="flex min-h-screen items-center justify-center">
				<div className="max-w-md text-center">
					<div className="mb-4 flex justify-center">
						<CheckCircle className="h-16 w-16 text-content-success" />
					</div>
					<h1 className="mb-2 text-2xl font-semibold">GitHub App Created!</h1>
					<p className="mb-4 text-content-secondary">
						Your GitHub App "{provider.display_name || provider.id}" has been
						successfully configured.
					</p>
					<p className="text-sm text-content-secondary">
						Redirecting to settings...
					</p>
				</div>
			</div>
		);
	}

	return null;
};
