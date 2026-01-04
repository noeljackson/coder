import {
	deleteWorkspaceCollaborator,
	deleteWorkspaceInvitation,
	updateWorkspaceCollaborator,
	workspaceCollaborators,
	workspaceInvitations,
} from "api/queries/workspaceInvitations";
import type {
	WorkspaceAccessLevel,
	WorkspaceCollaborator,
	WorkspaceInvitation,
} from "api/typesGenerated";
import { Avatar } from "components/Avatar/Avatar";
import { Button } from "components/Button/Button";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "components/Select/Select";
import { Spinner } from "components/Spinner/Spinner";
import { Clock, Mail, Trash2, UserPlus } from "lucide-react";
import { type FC, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { InviteCollaboratorDialog } from "./InviteCollaboratorDialog";

interface CollaboratorsPanelProps {
	workspaceId: string;
	workspaceName: string;
	isOwner: boolean;
}

const accessLevelLabels: Record<WorkspaceAccessLevel, string> = {
	readonly: "Read Only",
	use: "Use",
	admin: "Admin",
};

const formatExpiresAt = (expiresAt: string): string => {
	const date = new Date(expiresAt);
	const now = new Date();
	const diffMs = date.getTime() - now.getTime();
	const diffDays = Math.ceil(diffMs / (1000 * 60 * 60 * 24));

	if (diffDays < 0) return "Expired";
	if (diffDays === 0) return "Expires today";
	if (diffDays === 1) return "Expires tomorrow";
	return `Expires in ${diffDays} days`;
};

export const CollaboratorsPanel: FC<CollaboratorsPanelProps> = ({
	workspaceId,
	workspaceName,
	isOwner,
}) => {
	const queryClient = useQueryClient();
	const [isInviteDialogOpen, setIsInviteDialogOpen] = useState(false);

	const collaboratorsQuery = useQuery(workspaceCollaborators(workspaceId));
	const invitationsQuery = useQuery(workspaceInvitations(workspaceId));

	const removeCollaboratorMutation = useMutation(
		deleteWorkspaceCollaborator(queryClient),
	);
	const cancelInvitationMutation = useMutation(
		deleteWorkspaceInvitation(queryClient),
	);
	const updateAccessMutation = useMutation(
		updateWorkspaceCollaborator(queryClient),
	);

	const handleRemoveCollaborator = async (collaboratorId: string) => {
		await removeCollaboratorMutation.mutateAsync({
			workspaceId,
			collaboratorId,
		});
	};

	const handleCancelInvitation = async (invitationId: string) => {
		await cancelInvitationMutation.mutateAsync({
			workspaceId,
			invitationId,
		});
	};

	const handleUpdateAccess = async (
		collaboratorId: string,
		accessLevel: WorkspaceAccessLevel,
	) => {
		await updateAccessMutation.mutateAsync({
			workspaceId,
			collaboratorId,
			data: { access_level: accessLevel },
		});
	};

	const collaborators = collaboratorsQuery.data || [];
	const pendingInvitations = (invitationsQuery.data || []).filter(
		(inv) => inv.status === "pending",
	);

	const isLoading = collaboratorsQuery.isLoading || invitationsQuery.isLoading;

	return (
		<div className="flex flex-col gap-4">
			<div className="flex items-center justify-between">
				<h3 className="text-lg font-semibold text-content-primary">
					Collaborators
				</h3>
				{isOwner && (
					<Button
						variant="outline"
						size="sm"
						onClick={() => setIsInviteDialogOpen(true)}
					>
						<UserPlus className="size-4" />
						Invite
					</Button>
				)}
			</div>

			{isLoading ? (
				<div className="flex items-center justify-center py-8">
					<Spinner />
				</div>
			) : (
				<div className="flex flex-col gap-2">
					{collaborators.length === 0 && pendingInvitations.length === 0 ? (
						<p className="text-sm text-content-secondary py-4 text-center">
							No collaborators yet. Invite someone to get started.
						</p>
					) : (
						<>
							{collaborators.map((collaborator) => (
								<CollaboratorRow
									key={collaborator.id}
									collaborator={collaborator}
									isOwner={isOwner}
									onRemove={() => handleRemoveCollaborator(collaborator.id)}
									onUpdateAccess={(level) =>
										handleUpdateAccess(collaborator.id, level)
									}
									isUpdating={updateAccessMutation.isPending}
								/>
							))}

							{pendingInvitations.length > 0 && (
								<>
									<div className="border-t border-border my-2" />
									<p className="text-xs text-content-secondary uppercase tracking-wide font-medium">
										Pending Invitations
									</p>
									{pendingInvitations.map((invitation) => (
										<InvitationRow
											key={invitation.id}
											invitation={invitation}
											isOwner={isOwner}
											onCancel={() => handleCancelInvitation(invitation.id)}
										/>
									))}
								</>
							)}
						</>
					)}
				</div>
			)}

			<InviteCollaboratorDialog
				open={isInviteDialogOpen}
				onClose={() => setIsInviteDialogOpen(false)}
				workspaceId={workspaceId}
				workspaceName={workspaceName}
			/>
		</div>
	);
};

interface CollaboratorRowProps {
	collaborator: WorkspaceCollaborator;
	isOwner: boolean;
	onRemove: () => void;
	onUpdateAccess: (level: WorkspaceAccessLevel) => void;
	isUpdating: boolean;
}

const CollaboratorRow: FC<CollaboratorRowProps> = ({
	collaborator,
	isOwner,
	onRemove,
	onUpdateAccess,
	isUpdating,
}) => {
	return (
		<div className="flex items-center gap-3 p-3 rounded-md bg-surface-secondary">
			<Avatar
				src={collaborator.avatar_url}
				fallback={collaborator.username || collaborator.email}
			/>
			<div className="flex-1 min-w-0">
				<p className="text-sm font-medium text-content-primary truncate">
					{collaborator.username || "Unknown User"}
				</p>
				{collaborator.email && (
					<p className="text-xs text-content-secondary truncate">
						{collaborator.email}
					</p>
				)}
			</div>
			{isOwner ? (
				<div className="flex items-center gap-2">
					<Select
						value={collaborator.access_level}
						onValueChange={onUpdateAccess}
						disabled={isUpdating}
					>
						<SelectTrigger className="w-28">
							<SelectValue />
						</SelectTrigger>
						<SelectContent>
							{(["readonly", "use", "admin"] as WorkspaceAccessLevel[]).map(
								(level) => (
									<SelectItem key={level} value={level}>
										{accessLevelLabels[level]}
									</SelectItem>
								),
							)}
						</SelectContent>
					</Select>
					<Button
						variant="subtle"
						size="icon"
						onClick={onRemove}
						title="Remove collaborator"
					>
						<Trash2 className="size-4 text-content-destructive" />
					</Button>
				</div>
			) : (
				<span className="text-sm text-content-secondary">
					{accessLevelLabels[collaborator.access_level]}
				</span>
			)}
		</div>
	);
};

interface InvitationRowProps {
	invitation: WorkspaceInvitation;
	isOwner: boolean;
	onCancel: () => void;
}

const InvitationRow: FC<InvitationRowProps> = ({
	invitation,
	isOwner,
	onCancel,
}) => {
	return (
		<div className="flex items-center gap-3 p-3 rounded-md bg-surface-tertiary border border-dashed border-border">
			<div className="size-8 rounded-full bg-surface-secondary flex items-center justify-center">
				<Mail className="size-4 text-content-secondary" />
			</div>
			<div className="flex-1 min-w-0">
				<p className="text-sm font-medium text-content-primary truncate">
					{invitation.email}
				</p>
				<p className="text-xs text-content-secondary flex items-center gap-1">
					<Clock className="size-3" />
					{formatExpiresAt(invitation.expires_at)}
				</p>
			</div>
			<span className="text-xs text-content-secondary px-2 py-1 rounded bg-surface-secondary">
				{accessLevelLabels[invitation.access_level]}
			</span>
			{isOwner && (
				<Button
					variant="subtle"
					size="icon"
					onClick={onCancel}
					title="Cancel invitation"
				>
					<Trash2 className="size-4 text-content-destructive" />
				</Button>
			)}
		</div>
	);
};
