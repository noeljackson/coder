import { useMutation } from "@tanstack/react-query";
import { useState } from "react";
import { API } from "api/api";
import type { GitHubAppManifestRequest } from "api/typesGenerated";
import { Button } from "components/Button/Button";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
	DialogTrigger,
} from "components/Dialog/Dialog";
import { Input } from "components/Input/Input";
import { Label } from "components/Label/Label";
import { Spinner } from "components/Spinner/Spinner";
import { Github, Plus } from "lucide-react";

interface CreateGitHubAppDialogProps {
	onSuccess?: () => void;
}

export const CreateGitHubAppDialog: React.FC<
	CreateGitHubAppDialogProps
> = () => {
	const [open, setOpen] = useState(false);
	const [owner, setOwner] = useState("");

	const initiateManifest = useMutation({
		mutationFn: async (req: GitHubAppManifestRequest) => {
			return API.initiateGitHubAppManifest(req);
		},
		onSuccess: (data) => {
			// Redirect to GitHub to create the app.
			window.location.href = data.url;
		},
	});

	const handleSubmit = (e: React.FormEvent) => {
		e.preventDefault();
		// The redirect_uri should point to our callback page.
		const redirectUri = `${window.location.origin}/deployment/external-auth/github/callback`;
		initiateManifest.mutate({
			owner: owner.trim() || undefined,
			redirect_uri: redirectUri,
		});
	};

	return (
		<Dialog open={open} onOpenChange={setOpen}>
			<DialogTrigger asChild>
				<Button variant="outline">
					<Plus className="mr-2 h-4 w-4" />
					Create GitHub App
				</Button>
			</DialogTrigger>
			<DialogContent className="max-w-md">
				<DialogHeader>
					<div className="flex items-center gap-3">
						<div className="flex h-10 w-10 items-center justify-center rounded-lg bg-surface-secondary">
							<Github className="h-5 w-5" />
						</div>
						<div>
							<DialogTitle>Create GitHub App</DialogTitle>
							<DialogDescription>
								Create a GitHub App to enable Git authentication for your
								workspaces.
							</DialogDescription>
						</div>
					</div>
				</DialogHeader>

				<form onSubmit={handleSubmit}>
					<div className="space-y-4 py-4">
						<div className="space-y-2">
							<Label htmlFor="owner">Organization (optional)</Label>
							<Input
								id="owner"
								placeholder="Leave empty for personal account"
								value={owner}
								onChange={(e) => setOwner(e.target.value)}
								disabled={initiateManifest.isPending}
							/>
							<p className="text-xs text-content-secondary">
								Enter an organization name to create the app for that
								organization, or leave empty to create a personal app.
							</p>
						</div>
					</div>

					<DialogFooter>
						<Button
							type="button"
							variant="outline"
							onClick={() => setOpen(false)}
							disabled={initiateManifest.isPending}
						>
							Cancel
						</Button>
						<Button type="submit" disabled={initiateManifest.isPending}>
							{initiateManifest.isPending ? (
								<>
									<Spinner className="mr-2 h-4 w-4" />
									Redirecting...
								</>
							) : (
								"Continue to GitHub"
							)}
						</Button>
					</DialogFooter>
				</form>
			</DialogContent>
		</Dialog>
	);
};
