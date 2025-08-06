import { useMutation } from "@tanstack/react-query";

import { v1UserUpdateMutation } from "@/lib/client/@tanstack/react-query.gen";
import { showErrorToast, showSuccessToast } from "@/lib/toast";

export function useUpdatePassword() {
  return useMutation({
    ...v1UserUpdateMutation(),
    onSuccess: () => {
      showSuccessToast(
        "Password updated successfully",
        "Your password has been changed successfully"
      );
    },
    onError: (error) => {
      showErrorToast("Failed to update password", error.message);
    },
  });
}
