interface Message {
  id: number;
  title: string;
  message?: string;
  type: 'info' | 'success' | 'warning' | 'error';
  dismissAfter?: number;
}

type Drawers = {
  showTodos: boolean;
  showNotifications: boolean;
};
