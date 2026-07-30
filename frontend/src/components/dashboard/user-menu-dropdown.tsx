import { Link } from 'react-router-dom'

import { cn } from '@/lib/utils'

interface UserMenuDropdownProps {
  currentPath: string
  isLoggingOut?: boolean
  onClose: () => void
  onLogout: () => void
  placement?: 'above-start' | 'below-end'
}

const linkClassName =
  'block rounded-2xl px-4 py-3 text-sm text-[#5f5148] transition-colors hover:bg-[#f7f1ea] hover:text-[#2f221b]'

export function UserMenuDropdown({
  currentPath,
  isLoggingOut = false,
  onClose,
  onLogout,
  placement = 'below-end',
}: UserMenuDropdownProps) {
  return (
    <div
      className={cn(
        'absolute z-50 rounded-[28px] border border-[#eadfd5] bg-[#fcf8f3] p-3 shadow-[0_22px_50px_rgba(59,39,28,0.14)]',
        placement === 'above-start'
          ? 'bottom-[calc(100%+0.85rem)] left-0 w-full min-w-[12rem]'
          : 'right-0 top-[calc(100%+0.85rem)] w-[10rem]',
      )}
    >
      <div className="space-y-1">
        <Link
          className={cn(
            linkClassName,
            currentPath === '/profile' ? 'bg-[#f3ece4] font-medium text-[#2f221b]' : '',
          )}
          onClick={onClose}
          to="/profile"
        >
          个人资料
        </Link>
        {/* <Link className={linkClassName} onClick={onClose} to="/survey">
          修改问卷
        </Link> */}
        <Link
          className={cn(
            linkClassName,
            currentPath === '/feedback' ? 'bg-[#f3ece4] font-medium text-[#2f221b]' : '',
          )}
          onClick={onClose}
          to="/feedback"
        >
          反馈建议
        </Link>
      </div>

      <div className="mt-3 border-t border-[#eee2d8] pt-3">
        <button
          className="block w-full rounded-2xl px-4 py-3 text-left text-sm text-[#9e5e52] transition-colors hover:bg-[#fbf1ee] hover:text-[#7f4439]"
          disabled={isLoggingOut}
          onClick={onLogout}
          type="button"
        >
          {isLoggingOut ? '退出中...' : '退出登录'}
        </button>
      </div>
    </div>
  )
}
