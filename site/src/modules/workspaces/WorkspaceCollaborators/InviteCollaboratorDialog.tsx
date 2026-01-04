import { createWorkspaceInvitation } from "api/queries/workspaceInvitations";
import type {
	WorkspaceAccessLevel,
	WorkspaceInvitation,
} from "api/typesGenerated";
import { Button } from "components/Button/Button";
import {
	Dialog,
	DialogClose,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "components/Dialog/Dialog";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "components/Select/Select";
import { Spinner } from "components/Spinner/Spinner";
import { Mail, UserPlus } from "lucide-react";
import { type FC, useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";

interface InviteCollaboratorDialogProps {
	open: boolean;
	onClose: () => void;
	workspaceId: string;
	workspaceName: string;
	onSuccess?: (invitation: WorkspaceInvitation) => void;
}

const accessLevelLabels: Record<WorkspaceAccessLevel, string> = {
	readonly: "Read Only",
	use: "Use",
	admin: "Admin",
};

const accessLevelDescriptions: Record<WorkspaceAccessLevel, string> = {
	readonly: "Can view workspace status and logs",
	use: "Can connect and use the workspace",
	admin: "Full access including settings and invitations",
};

export const InviteCollaboratorDialog: FC<InviteCollaboratorDialogProps> = ({
	open,
	onClose,
	workspaceId,
	workspaceName,
	onSuccess,
}) => {
	const queryClient = useQueryClient();
	const [email, setEmail] = useState("");
	const [accessLevel, setAccessLevel] = useState<WorkspaceAccessLevel>("use");
	const [error, setError] = useState<string | null>(null);

	const inviteMutation = useMutation(createWorkspaceInvitation(queryClient));

	const handleSubmit = async (e: React.FormEvent) => {
		e.preventDefault();
		setError(null);

		if (!email || !email.includes("@")) {
			setError("Please enter a valid email address");
			return;
		}

		try {
			const invitation = await inviteMutation.mutateAsync({
				workspaceId,
				data: {
					email,
					access_level: accessLevel,
				},
			});

			setEmail("");
			setAccessLevel("use");
			onSuccess?.(invitation);
			onClose();
		} catch (err) {
			if (err instanceof Error) {
				setError(err.message);
			} else {
				setError("Failed to send invitation");
			}
		}
	};

	const handleOpenChange = (isOpen: boolean) => {
		if (!isOpen) {
			setEmail("");
			setAccessLevel("use");
			setError(null);
			onClose();
		}
	};

	return (
		<Dialog open={open} onOpenChange={handleOpenChange}>
			<DialogContent>
				<DialogHeader>
					<DialogTitle className="flex items-center gap-2">
						<UserPlus className="size-5" />
						Invite Collaborator
					</DialogTitle>
					<DialogDescription>
						Invite someone to collaborate on <strong>{workspaceName}</strong>.
						They will receive an email with a link to accept the invitation.
					</DialogDescription>
				</DialogHeader>

				<form onSubmit={handleSubmit} className="flex flex-col gap-4">
					<div className="flex flex-col gap-2">
						<label
							htmlFor="email"
							className="text-sm font-medium text-content-primary"
						>
							Email Address
						</label>
						<div className="relative">
							<Mail className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-content-secondary" />
							<input
								id="email"
								type="email"
								value={email}
								onChange={(e) => setEmail(e.target.value)}
								placeholder="collaborator@example.com"
								className="w-full h-10 pl-10 pr-3 rounded-md border border-border border-solid
									bg-transparent text-sm text-content-primary
									placeholder:text-content-secondary
									focus:outline-none focus:ring-2 focus:ring-content-link"
								disabled={inviteMutation.isPending}
							/>
						</div>
					</div>

					<div className="flex flex-col gap-2">
						<label
							htmlFor="access-level"
							className="text-sm font-medium text-content-primary"
						>
							Access Level
						</label>
						<Select
							value={accessLevel}
							onValueChange={(value) =>
								setAccessLevel(value as WorkspaceAccessLevel)
							}
							disabled={inviteMutation.isPending}
						>
							<SelectTrigger id="access-level">
								<SelectValue placeholder="Select access level" />
							</SelectTrigger>
							<SelectContent>
								{(["readonly", "use", "admin"] as WorkspaceAccessLevel[]).map(
									(level) => (
										<SelectItem key={level} value={level}>
											<div className="flex flex-col">
												<span className="font-medium">
													{accessLevelLabels[level]}
												</span>
												<span className="text-xs text-content-secondary">
													{accessLevelDescriptions[level]}
												</span>
											</div>
										</SelectItem>
									),
								)}
							</SelectContent>
						</Select>
					</div>

					{error && (
						<p className="text-sm text-content-destructive bg-surface-destructive p-3 rounded-md">
							{error}
						</p>
					)}

					<DialogFooter className="mt-2">
						<DialogClose asChild>
							<Button variant="outline" disabled={inviteMutation.isPending}>
								Cancel
							</Button>
						</DialogClose>
						<Button type="submit" disabled={inviteMutation.isPending}>
							{inviteMutation.isPending ? (
								<>
									<Spinner size="sm" />
									Sending...
								</>
							) : (
								"Send Invitation"
							)}
						</Button>
					</DialogFooter>
				</form>
			</DialogContent>
		</Dialog>
	);
};
