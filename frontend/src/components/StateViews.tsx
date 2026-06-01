import { AlertCircle, Loader2, SearchX } from "lucide-react";

export function LoadingState({ label = "正在加载..." }: { label?: string }) {
  return (
    <div className="state-box">
      <Loader2 className="spin" size={22} aria-hidden="true" />
      <div className="state-copy">
        <strong>请稍候</strong>
        <span>{label}</span>
      </div>
    </div>
  );
}

export function ErrorState({ message, onRetry }: { message: string; onRetry?: () => void }) {
  return (
    <div className="state-box error-state">
      <AlertCircle size={22} aria-hidden="true" />
      <div className="state-copy">
        <strong>发生错误</strong>
        <span>{message}</span>
      </div>
      {onRetry ? <button className="ghost-button compact" type="button" onClick={onRetry}>重试</button> : null}
    </div>
  );
}

export function EmptyState({ title, description }: { title: string; description: string }) {
  return (
    <div className="state-box empty-state">
      <SearchX size={24} aria-hidden="true" />
      <div className="state-copy">
        <strong>{title}</strong>
        <span>{description}</span>
      </div>
    </div>
  );
}
