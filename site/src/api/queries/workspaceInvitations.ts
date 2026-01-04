import { API } from "api/api";
import type {
	CreateWorkspaceInvitationRequest,
	UpdateWorkspaceCollaboratorRequest,
	WorkspaceCollaborator,
	WorkspaceInvitation,
} from "api/typesGenerated";
import type { MutationOptions, QueryClient, QueryOptions } from "react-query";

// Query keys
export const workspaceInvitationsKey = (workspaceId: string) => [
	"workspaces",
	workspaceId,
	"invitations",
];

export const workspaceCollaboratorsKey = (workspaceId: string) => [
	"workspaces",
	workspaceId,
	"collaborators",
];

export const invitationByTokenKey = (token: string) => ["invitations", token];

export const myPendingInvitationsKey = () => ["users", "me", "invitations"];

export const myCollaborationsKey = () => ["users", "me", "collaborations"];

// Queries
export const workspaceInvitations = (workspaceId: string) => {
	return {
		queryKey: workspaceInvitationsKey(workspaceId),
		queryFn: () => API.getWorkspaceInvitations(workspaceId),
	} satisfies QueryOptions<WorkspaceInvitation[]>;
};

export const workspaceCollaborators = (workspaceId: string) => {
	return {
		queryKey: workspaceCollaboratorsKey(workspaceId),
		queryFn: () => API.getWorkspaceCollaborators(workspaceId),
	} satisfies QueryOptions<WorkspaceCollaborator[]>;
};

export const invitationByToken = (token: string) => {
	return {
		queryKey: invitationByTokenKey(token),
		queryFn: () => API.getWorkspaceInvitationByToken(token),
	} satisfies QueryOptions<WorkspaceInvitation>;
};

export const myPendingInvitations = () => {
	return {
		queryKey: myPendingInvitationsKey(),
		queryFn: () => API.getMyPendingInvitations(),
	} satisfies QueryOptions<WorkspaceInvitation[]>;
};

export const myCollaborations = () => {
	return {
		queryKey: myCollaborationsKey(),
		queryFn: () => API.getMyWorkspaceCollaborations(),
	} satisfies QueryOptions<WorkspaceCollaborator[]>;
};

// Mutations
export const createWorkspaceInvitation = (
	queryClient: QueryClient,
): MutationOptions<
	WorkspaceInvitation,
	unknown,
	{ workspaceId: string; data: CreateWorkspaceInvitationRequest }
> => {
	return {
		mutationFn: ({ workspaceId, data }) =>
			API.createWorkspaceInvitation(workspaceId, data),
		onSuccess: async (_res, { workspaceId }) => {
			await queryClient.invalidateQueries({
				queryKey: workspaceInvitationsKey(workspaceId),
			});
		},
	};
};

export const deleteWorkspaceInvitation = (
	queryClient: QueryClient,
): MutationOptions<
	void,
	unknown,
	{ workspaceId: string; invitationId: string }
> => {
	return {
		mutationFn: ({ workspaceId, invitationId }) =>
			API.deleteWorkspaceInvitation(workspaceId, invitationId),
		onSuccess: async (_res, { workspaceId }) => {
			await queryClient.invalidateQueries({
				queryKey: workspaceInvitationsKey(workspaceId),
			});
		},
	};
};

export const acceptWorkspaceInvitation = (
	queryClient: QueryClient,
): MutationOptions<WorkspaceCollaborator, unknown, { token: string }> => {
	return {
		mutationFn: ({ token }) => API.acceptWorkspaceInvitation(token),
		onSuccess: async () => {
			await queryClient.invalidateQueries({
				queryKey: myPendingInvitationsKey(),
			});
			await queryClient.invalidateQueries({
				queryKey: myCollaborationsKey(),
			});
		},
	};
};

export const declineWorkspaceInvitation = (
	queryClient: QueryClient,
): MutationOptions<void, unknown, { token: string }> => {
	return {
		mutationFn: ({ token }) => API.declineWorkspaceInvitation(token),
		onSuccess: async () => {
			await queryClient.invalidateQueries({
				queryKey: myPendingInvitationsKey(),
			});
		},
	};
};

export const updateWorkspaceCollaborator = (
	queryClient: QueryClient,
): MutationOptions<
	WorkspaceCollaborator,
	unknown,
	{
		workspaceId: string;
		collaboratorId: string;
		data: UpdateWorkspaceCollaboratorRequest;
	}
> => {
	return {
		mutationFn: ({ workspaceId, collaboratorId, data }) =>
			API.updateWorkspaceCollaborator(workspaceId, collaboratorId, data),
		onSuccess: async (_res, { workspaceId }) => {
			await queryClient.invalidateQueries({
				queryKey: workspaceCollaboratorsKey(workspaceId),
			});
		},
	};
};

export const deleteWorkspaceCollaborator = (
	queryClient: QueryClient,
): MutationOptions<
	void,
	unknown,
	{ workspaceId: string; collaboratorId: string }
> => {
	return {
		mutationFn: ({ workspaceId, collaboratorId }) =>
			API.deleteWorkspaceCollaborator(workspaceId, collaboratorId),
		onSuccess: async (_res, { workspaceId }) => {
			await queryClient.invalidateQueries({
				queryKey: workspaceCollaboratorsKey(workspaceId),
			});
		},
	};
};
