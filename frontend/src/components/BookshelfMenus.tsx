import { MoreHorizontal } from "lucide-react";
import type { ReactNode } from "react";
import { useEffect, useRef, useState } from "react";
import type { BookshelfEntry, BookshelfGroup } from "../api/types";

const NEW_GROUP_VALUE = "__new__";

function ActionMenu({
  label,
  onOpenChange,
  children,
}: {
  label: string;
  onOpenChange?: (nextOpen: boolean) => void;
  children: (controls: { close: () => void }) => ReactNode;
}) {
  const [isOpen, setIsOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement | null>(null);

  function updateOpen(nextOpen: boolean) {
    setIsOpen(nextOpen);
    onOpenChange?.(nextOpen);
  }

  useEffect(() => {
    if (!isOpen) {
      return;
    }
    function handlePointerDown(event: MouseEvent) {
      if (!containerRef.current?.contains(event.target as Node)) {
        updateOpen(false);
      }
    }
    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape") {
        updateOpen(false);
      }
    }
    document.addEventListener("mousedown", handlePointerDown);
    document.addEventListener("keydown", handleKeyDown);
    return () => {
      document.removeEventListener("mousedown", handlePointerDown);
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, [isOpen]);

  return (
    <div ref={containerRef} className="bookshelf-card-menu-shell">
      <button
        className="bookshelf-card-menu-trigger"
        type="button"
        aria-label={label}
        aria-haspopup="menu"
        aria-expanded={isOpen}
        onClick={() => updateOpen(!isOpen)}
      >
        <MoreHorizontal size={18} aria-hidden="true" />
      </button>
      {isOpen ? (
        <div className="bookshelf-card-menu" role="menu">
          {children({ close: () => updateOpen(false) })}
        </div>
      ) : null}
    </div>
  );
}

export function BookshelfGroupMenu({
  group,
  onTogglePin,
  onRename,
  onDelete,
}: {
  group: BookshelfGroup;
  onTogglePin: (group: BookshelfGroup) => Promise<void> | void;
  onRename: (group: BookshelfGroup) => Promise<void> | void;
  onDelete: (group: BookshelfGroup) => Promise<void> | void;
}) {
  return (
    <ActionMenu label={`管理分组“${group.name}”`}>
      {({ close }) => (
        <>
          <button
            className="bookshelf-card-menu-item"
            type="button"
            onClick={() => {
              close();
              void onTogglePin(group);
            }}
          >
            {group.isPinned ? "取消置顶" : "置顶"}
          </button>
          <button
            className="bookshelf-card-menu-item"
            type="button"
            onClick={() => {
              close();
              void onRename(group);
            }}
          >
            重命名
          </button>
          <button
            className="bookshelf-card-menu-item danger"
            type="button"
            onClick={() => {
              close();
              void onDelete(group);
            }}
          >
            删除
          </button>
        </>
      )}
    </ActionMenu>
  );
}

export function BookshelfRenameGroupDialog({
  group,
  onClose,
  onSubmit,
}: {
  group: BookshelfGroup | null;
  onClose: () => void;
  onSubmit: (group: BookshelfGroup, nextName: string) => Promise<void> | void;
}) {
  const [nextName, setNextName] = useState(group?.name ?? "");

  useEffect(() => {
    setNextName(group?.name ?? "");
  }, [group]);

  if (!group) {
    return null;
  }

  return (
    <div className="modal-backdrop" role="dialog" aria-modal="true" aria-label={`重命名分组“${group.name}”`}>
      <div className="modal panel">
        <div className="modal-head">
          <div>
            <p className="eyebrow">分组设置</p>
            <h2>重命名分组</h2>
          </div>
          <button className="ghost-button compact" type="button" onClick={onClose}>关闭</button>
        </div>
        <div className="bookshelf-move-dialog">
          <label>
            分组名称
            <input
              value={nextName}
              maxLength={64}
              placeholder="输入新的分组名称"
              onChange={(event) => setNextName(event.target.value)}
            />
          </label>
        </div>
        <div className="modal-actions">
          <button className="ghost-button compact" type="button" onClick={onClose}>取消</button>
          <button
            className="primary-button compact"
            type="button"
            disabled={!nextName.trim()}
            onClick={() => void onSubmit(group, nextName.trim())}
          >
            保存
          </button>
        </div>
      </div>
    </div>
  );
}

export function BookshelfBookMenu({
  entry,
  onTogglePin,
  onRequestMove,
  onRemove,
}: {
  entry: BookshelfEntry;
  onTogglePin: (entry: BookshelfEntry) => Promise<void> | void;
  onRequestMove: (entry: BookshelfEntry) => void;
  onRemove: (bookId: number) => Promise<void> | void;
}) {
  return (
    <ActionMenu label={`管理《${entry.book.title}》`}>
      {({ close }) => (
        <>
          <button
            className="bookshelf-card-menu-item"
            type="button"
            onClick={() => {
              close();
              void onTogglePin(entry);
            }}
          >
            {entry.isPinned ? "取消置顶" : "置顶"}
          </button>
          <button
            className="bookshelf-card-menu-item"
            type="button"
            onClick={() => {
              close();
              onRequestMove(entry);
            }}
          >
            移动
          </button>
          <button
            className="bookshelf-card-menu-item danger"
            type="button"
            onClick={() => {
              close();
              void onRemove(entry.bookId);
            }}
          >
            移出
          </button>
        </>
      )}
    </ActionMenu>
  );
}

export function BookshelfMoveDialog({
  entry,
  groups,
  currentGroupId,
  allowMoveToShelf = false,
  onClose,
  onMove,
  onCreateGroupAndMove,
}: {
  entry: BookshelfEntry | null;
  groups: BookshelfGroup[];
  currentGroupId?: number;
  allowMoveToShelf?: boolean;
  onClose: () => void;
  onMove: (bookId: number, groupId: number) => Promise<void> | void;
  onCreateGroupAndMove: (bookId: number, groupName: string) => Promise<void> | void;
}) {
  const availableGroups = groups.filter((group) => group.id !== currentGroupId);

  function defaultTargetValue() {
    if (allowMoveToShelf) {
      return "0";
    }
    if (availableGroups.length > 0) {
      return String(availableGroups[0].id);
    }
    return NEW_GROUP_VALUE;
  }

  const [targetGroupValue, setTargetGroupValue] = useState(defaultTargetValue);
  const [newGroupName, setNewGroupName] = useState("");

  useEffect(() => {
    if (!entry) {
      return;
    }
    setTargetGroupValue(defaultTargetValue());
    setNewGroupName("");
  }, [entry, allowMoveToShelf, groups, currentGroupId]);

  async function handleConfirm() {
    if (!entry) {
      return;
    }
    if (targetGroupValue === NEW_GROUP_VALUE) {
      const nextName = newGroupName.trim();
      if (!nextName) {
        return;
      }
      await onCreateGroupAndMove(entry.bookId, nextName);
      onClose();
      return;
    }
    await onMove(entry.bookId, Number(targetGroupValue));
    onClose();
  }

  if (!entry) {
    return null;
  }

  return (
    <div className="modal-backdrop" role="dialog" aria-modal="true" aria-label={`移动《${entry.book.title}》`}>
      <div className="modal panel">
        <div className="modal-head">
          <div>
            <p className="eyebrow">移动书籍</p>
            <h2>《{entry.book.title}》</h2>
          </div>
          <button className="ghost-button compact" type="button" onClick={onClose}>关闭</button>
        </div>
        <div className="bookshelf-move-dialog">
          <label>
            目标位置
            <select value={targetGroupValue} onChange={(event) => setTargetGroupValue(event.target.value)}>
              {allowMoveToShelf ? <option value="0">书架首页</option> : null}
              {availableGroups.map((group) => (
                <option key={group.id} value={group.id}>{group.name}</option>
              ))}
              <option value={NEW_GROUP_VALUE}>新建分组</option>
            </select>
          </label>
          {targetGroupValue === NEW_GROUP_VALUE ? (
            <label>
              分组名称
              <input
                value={newGroupName}
                maxLength={64}
                placeholder="例如：周末想看"
                onChange={(event) => setNewGroupName(event.target.value)}
              />
            </label>
          ) : null}
        </div>
        <div className="modal-actions">
          <button className="ghost-button compact" type="button" onClick={onClose}>取消</button>
          <button
            className="primary-button compact"
            type="button"
            disabled={targetGroupValue === NEW_GROUP_VALUE && !newGroupName.trim()}
            onClick={() => void handleConfirm()}
          >
            确认移动
          </button>
        </div>
      </div>
    </div>
  );
}
