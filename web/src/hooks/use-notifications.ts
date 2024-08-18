import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import {
  v1NotificationDeleteMutation,
  v1NotificationsGetOptions,
} from "@/lib/api";
import { showErrorToast, showSuccessToast } from "@/lib/toast";

export function useNotifications() {
  return useQuery({
    ...v1NotificationsGetOptions(),
  });
}

export function useDeleteNotification() {
  const queryClient = useQueryClient();

  return useMutation({
    ...v1NotificationDeleteMutation(),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: v1NotificationsGetOptions().queryKey,
      });
      showSuccessToast(
        "Notification deleted",
        "The notification has been removed"
      );
    },
    onError: (error) => {
      showErrorToast("Failed to delete notification", error.message);
    },
  });
}
