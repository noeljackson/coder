import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { CheckCircle, Mail, Users, XCircle } from "lucide-react";
import { type FC, useState } from "react";
import { useNavigate, useParams } from "react-router";
import {
	acceptWorkspaceInvitation,
	declineWorkspaceInvitation,
	invitationByToken,
} from "api/queries/workspaceInvitations";
import type { WorkspaceAccessLevel } from "api/typesGenerated";
import { Button } from "components/Button/Button";
import { Loader } from "components/Loader/Loader";
import { Margins } from "components/Margins/Margins";
import { Spinner } from "components/Spinner/Spinner";

const accessLevelLabels: Record<WorkspaceAccessLevel, string> = {
	readonly: "Read Only",
	use: "Use",
	admin: "Admin",
};

const accessLevelDescriptions: Record<WorkspaceAccessLevel, string> = {
	readonly: "View workspace status and logs",
	use: "Connect and use the workspace",
	admin: "Full access including settings and invitations",
};

export const InvitationPage: FC = () => {
	const { token } = useParams<{ token: string }>();
	const navigate = useNavigate();
	const queryClient = useQueryClient();
	const [actionTaken, setActionTaken] = useState<
		"accepted" | "declined" | null
	>(null);

	const invitationQuery = useQuery({
		...invitationByToken(token || ""),
		enabled: !!token,
		retry: false,
	});

	const acceptMutation = useMutation(acceptWorkspaceInvitation(queryClient));
	const declineMutation = useMutation(declineWorkspaceInvitation(queryClient));

	const handleAccept = async () => {
		if (!token) return;
		try {
			await acceptMutation.mutateAsync({ token });
			setActionTaken("accepted");
		} catch (err) {
			console.error("Failed to accept invitation:", err);
		}
	};

	const handleDecline = async () => {
		if (!token) return;
		try {
			await declineMutation.mutateAsync({ token });
			setActionTaken("declined");
		} catch (err) {
			console.error("Failed to decline invitation:", err);
		}
	};

	if (invitationQuery.isLoading) {
		return <Loader />;
	}

	if (invitationQuery.error || !invitationQuery.data) {
		return (
			<Margins>
				<div className="flex flex-col items-center justify-center min-h-[60vh] gap-6">
					<div className="size-16 rounded-full bg-surface-destructive flex items-center justify-center">
						<XCircle className="size-8 text-content-destructive" />
					</div>
					<div className="text-center">
						<h1 className="text-2xl font-semibold text-content-primary mb-2">
							Invitation Not Found
						</h1>
						<p className="text-content-secondary max-w-md">
							This invitation may have expired, been canceled, or already been
							used. Please contact the workspace owner for a new invitation.
						</p>
					</div>
					<Button onClick={() => navigate("/workspaces")}>
						Go to Workspaces
					</Button>
				</div>
			</Margins>
		);
	}

	const invitation = invitationQuery.data;

	if (actionTaken === "accepted") {
		return (
			<Margins>
				<div className="flex flex-col items-center justify-center min-h-[60vh] gap-6">
					<div className="size-16 rounded-full bg-green-500/10 flex items-center justify-center">
						<CheckCircle className="size-8 text-green-500" />
					</div>
					<div className="text-center">
						<h1 className="text-2xl font-semibold text-content-primary mb-2">
							Invitation Accepted
						</h1>
						<p className="text-content-secondary max-w-md">
							You now have{" "}
							<strong>{accessLevelLabels[invitation.access_level]}</strong>{" "}
							access to <strong>{invitation.workspace_name}</strong>.
						</p>
					</div>
					<Button onClick={() => navigate("/workspaces")}>
						Go to Workspaces
					</Button>
				</div>
			</Margins>
		);
	}

	if (actionTaken === "declined") {
		return (
			<Margins>
				<div className="flex flex-col items-center justify-center min-h-[60vh] gap-6">
					<div className="size-16 rounded-full bg-surface-secondary flex items-center justify-center">
						<XCircle className="size-8 text-content-secondary" />
					</div>
					<div className="text-center">
						<h1 className="text-2xl font-semibold text-content-primary mb-2">
							Invitation Declined
						</h1>
						<p className="text-content-secondary max-w-md">
							You have declined the invitation to collaborate on{" "}
							<strong>{invitation.workspace_name}</strong>.
						</p>
					</div>
					<Button onClick={() => navigate("/workspaces")}>
						Go to Workspaces
					</Button>
				</div>
			</Margins>
		);
	}

	if (invitation.status !== "pending") {
		return (
			<Margins>
				<div className="flex flex-col items-center justify-center min-h-[60vh] gap-6">
					<div className="size-16 rounded-full bg-surface-secondary flex items-center justify-center">
						<Mail className="size-8 text-content-secondary" />
					</div>
					<div className="text-center">
						<h1 className="text-2xl font-semibold text-content-primary mb-2">
							Invitation{" "}
							{invitation.status.charAt(0).toUpperCase() +
								invitation.status.slice(1)}
						</h1>
						<p className="text-content-secondary max-w-md">
							This invitation has already been {invitation.status}.
						</p>
					</div>
					<Button onClick={() => navigate("/workspaces")}>
						Go to Workspaces
					</Button>
				</div>
			</Margins>
		);
	}

	const isExpired = new Date(invitation.expires_at) < new Date();
	if (isExpired) {
		return (
			<Margins>
				<div className="flex flex-col items-center justify-center min-h-[60vh] gap-6">
					<div className="size-16 rounded-full bg-surface-destructive flex items-center justify-center">
						<XCircle className="size-8 text-content-destructive" />
					</div>
					<div className="text-center">
						<h1 className="text-2xl font-semibold text-content-primary mb-2">
							Invitation Expired
						</h1>
						<p className="text-content-secondary max-w-md">
							This invitation has expired. Please contact the workspace owner
							for a new invitation.
						</p>
					</div>
					<Button onClick={() => navigate("/workspaces")}>
						Go to Workspaces
					</Button>
				</div>
			</Margins>
		);
	}

	const isPending = acceptMutation.isPending || declineMutation.isPending;

	return (
		<Margins>
			<div className="flex flex-col items-center justify-center min-h-[60vh] gap-6">
				<div className="size-16 rounded-full bg-surface-secondary flex items-center justify-center">
					<Users className="size-8 text-content-link" />
				</div>

				<div className="text-center max-w-lg">
					<h1 className="text-2xl font-semibold text-content-primary mb-2">
						Workspace Invitation
					</h1>
					<p className="text-content-secondary">
						<strong>{invitation.inviter_username}</strong> has invited you to
						collaborate on <strong>{invitation.workspace_name}</strong>
					</p>
				</div>

				<div className="bg-surface-secondary rounded-lg p-6 w-full max-w-md">
					<div className="flex flex-col gap-4">
						<div className="flex items-center justify-between">
							<span className="text-sm text-content-secondary">
								Access Level
							</span>
							<span className="text-sm font-medium text-content-primary">
								{accessLevelLabels[invitation.access_level]}
							</span>
						</div>
						<p className="text-sm text-content-secondary bg-surface-tertiary p-3 rounded">
							{accessLevelDescriptions[invitation.access_level]}
						</p>
						<div className="flex items-center justify-between text-sm">
							<span className="text-content-secondary">Expires</span>
							<span className="text-content-primary">
								{new Date(invitation.expires_at).toLocaleDateString()}
							</span>
						</div>
					</div>
				</div>

				<div className="flex gap-3">
					<Button
						variant="outline"
						onClick={handleDecline}
						disabled={isPending}
					>
						{declineMutation.isPending ? <Spinner size="sm" /> : null}
						Decline
					</Button>
					<Button onClick={handleAccept} disabled={isPending}>
						{acceptMutation.isPending ? <Spinner size="sm" /> : null}
						Accept Invitation
					</Button>
				</div>

				{(acceptMutation.error ?? declineMutation.error) ? (
					<p className="text-sm text-content-destructive bg-surface-destructive p-3 rounded-md max-w-md text-center">
						{acceptMutation.error instanceof Error
							? acceptMutation.error.message
							: declineMutation.error instanceof Error
								? declineMutation.error.message
								: "An error occurred. Please try again."}
					</p>
				) : null}
			</div>
		</Margins>
	);
};

export default InvitationPage;
