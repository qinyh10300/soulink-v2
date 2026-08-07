import {
  ClipboardList,
  CalendarDays,
  CircleDotDashed,
  Clock3,
  type LucideIcon,
  LogOut,
  Menu,
  MessageCircleMore,
  MessagesSquare,
  PanelLeftClose,
  Plus,
  RotateCcw,
  Share2,
  ShieldCheck,
  Sparkles,
  UserRound,
  X,
} from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { Group, Panel, usePanelRef } from "react-resizable-panels";
import {
  Link,
  NavLink,
  Outlet,
  useLocation,
  useNavigate,
} from "react-router-dom";

import logoUrl from "@/assets/soulink-logo-icon.png";
import { ConnectionJourneyNav } from "@/components/connection/connection-journey-nav";
import { NotificationCenter } from "@/components/engagement/notification-center";
import { AuthenticatedImage } from "@/components/media/authenticated-image";
import { ProfileShareDrawer } from "@/components/profile/profile-share-drawer";
import { OnboardingDialog } from "@/components/engagement/onboarding-dialog";
import { PWAUpdateBanner } from "@/components/engagement/pwa-update-banner";
import { logout } from "@/lib/auth-api";
import { fetchMyIntents, type ConnectionIntent } from "@/lib/connection-api";
import {
  connectionIntentPath,
  connectionPaths,
  isConnectionPath,
} from "@/lib/connection-routes";
import { cn } from "@/lib/utils";
import { useChat } from "@/contexts/chat-context";
import { useMediaQuery } from "@/hooks/use-media-query";
import { usePersistentPanelLayout } from "@/hooks/use-persistent-panel-layout";
import {
  clearAuthSession,
  clearUserProfile,
  getAuthSession,
  getStoredUserProfile,
} from "@/lib/user-profile-storage";
import { WorkspaceEdgeButton } from "@/components/workspace/workspace-edge-button";
import { WorkspaceResizeHandle } from "@/components/workspace/workspace-resize-handle";

interface NavItem {
  label: string;
  mobileLabel: string;
  to: string;
  icon: LucideIcon;
  match: (pathname: string) => boolean;
}

const navItems: NavItem[] = [
  {
    label: "认识",
    mobileLabel: "认识",
    to: connectionPaths.root,
    icon: Sparkles,
    match: isConnectionPath,
  },
  {
    label: "广场",
    mobileLabel: "广场",
    to: "/plaza",
    icon: CircleDotDashed,
    match: (pathname) => pathname === "/plaza",
  },
  {
    label: "时间表",
    mobileLabel: "时间",
    to: "/calendar",
    icon: CalendarDays,
    match: (pathname) => pathname === "/calendar",
  },
  {
    label: "消息",
    mobileLabel: "消息",
    to: "/chat",
    icon: MessageCircleMore,
    match: (pathname) => pathname.startsWith("/chat"),
  },
  {
    label: "我的",
    mobileLabel: "我的",
    to: "/profile",
    icon: UserRound,
    match: (pathname) =>
      pathname === "/profile" ||
      pathname === "/account",
  },
];

function Brand() {
  return (
    <Link className="flex min-w-0 items-center gap-2.5" to={connectionPaths.root}>
      <img alt="" className="h-8 w-8 rounded-md object-cover" src={logoUrl} />
      <span className="min-w-0">
        <span className="block truncate text-[17px] font-semibold text-[#211f1d]">
          SouLink
        </span>
        <span className="block truncate text-[10px] font-semibold text-[#6c776f]">
          TSINGHUA · BETA
        </span>
      </span>
    </Link>
  );
}

export function ConnectionShell() {
  const navigate = useNavigate();
  const location = useLocation();
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const [loggingOut, setLoggingOut] = useState(false);
  const [shareDrawerOpen, setShareDrawerOpen] = useState(false);
  const [recentIntents, setRecentIntents] = useState<ConnectionIntent[]>([]);
  const { totalUnread } = useChat();
  const profile = useMemo(() => getStoredUserProfile(), []);
  const session = useMemo(() => getAuthSession(), []);
  const isDesktop = useMediaQuery("(min-width: 1024px)");
  const navigationPanelRef = usePanelRef();
  const navigationLayout = usePersistentPanelLayout(
    "soulink-primary-workbench-pixels-v2",
    {
      navigation: {
        defaultSize: 244,
        minSize: 208,
        maxSize: 336,
        rememberSize: true,
      },
      content: {},
    },
  );
  const [navigationCollapsed, setNavigationCollapsed] = useState(() =>
    navigationLayout.isInitiallyCollapsed("navigation"),
  );
  const inConnectionJourney = isConnectionPath(location.pathname);
  const inAdminWorkspace = location.pathname.startsWith("/admin");

  useEffect(() => {
    let cancelled = false;
    void fetchMyIntents()
      .then((intents) => {
        if (!cancelled) {
          setRecentIntents(
            [...intents]
              .sort(
                (left, right) =>
                  new Date(right.updatedAt).getTime() -
                  new Date(left.updatedAt).getTime(),
              )
              .slice(0, 4),
          );
        }
      })
      .catch(() => {
        if (!cancelled) {
          setRecentIntents([]);
        }
      });

    return () => {
      cancelled = true;
    };
  }, [location.pathname]);

  const handleLogout = async () => {
    if (loggingOut) {
      return;
    }
    setLoggingOut(true);
    try {
      await logout();
    } catch {
      // Local credentials are cleared even when the network request fails.
    } finally {
      clearAuthSession();
      clearUserProfile();
      navigate("/login");
    }
  };

  const resetNavigationWidth = () => {
    navigationLayout.reset();
    navigationPanelRef.current?.expand();
    navigationPanelRef.current?.resize(244);
  };

  return (
    <div className="min-h-screen bg-[#f7f6f3] text-[#23211f]">
      {!inAdminWorkspace ? <OnboardingDialog /> : null}
      <PWAUpdateBanner />
      <Group
        className="min-h-screen w-full"
        elementRef={navigationLayout.groupElementRef}
        id="soulink-primary-workbench"
        onLayoutChanged={
          isDesktop ? navigationLayout.onLayoutChanged : undefined
        }
        orientation="horizontal"
        resizeTargetMinimumSize={{ coarse: 24, fine: 11 }}
      >
        {isDesktop ? (
          <>
            <Panel
              collapsedSize={0}
              collapsible
              defaultSize={navigationLayout.getDefaultSize("navigation")}
              groupResizeBehavior="preserve-pixel-size"
              id="navigation"
              maxSize={336}
              minSize={208}
              onResize={({ inPixels }) => setNavigationCollapsed(inPixels < 1)}
              panelRef={navigationPanelRef}
            >
              <aside className="flex h-screen w-full flex-col overflow-hidden bg-[#fbfaf8] px-4 py-5">
                <div className="flex items-center justify-between gap-2">
                  <Brand />
                  <div className="flex shrink-0 items-center gap-0.5">
                    <NotificationCenter />
                    <button
                      aria-label="恢复默认侧栏宽度"
                      className="inline-flex h-8 w-8 items-center justify-center rounded-md text-[#8a857e] transition-colors hover:bg-[#efede9] hover:text-[#445f55] focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-[#527b6d]"
                      onClick={resetNavigationWidth}
                      title="恢复默认侧栏宽度"
                      type="button"
                    >
                      <RotateCcw size={14} />
                    </button>
                    <button
                      aria-label="隐藏主侧栏"
                      className="inline-flex h-8 w-8 items-center justify-center rounded-md text-[#8a857e] transition-colors hover:bg-[#efede9] hover:text-[#445f55] focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-[#527b6d]"
                      onClick={() => navigationPanelRef.current?.collapse()}
                      title="隐藏主侧栏"
                      type="button"
                    >
                      <PanelLeftClose size={15} />
                    </button>
                  </div>
                </div>

                <Link
                  className="mt-7 inline-flex h-10 shrink-0 items-center justify-center gap-2 rounded-md bg-[#263b35] px-4 text-sm font-semibold text-white transition-colors hover:bg-[#1d302a]"
                  to={connectionPaths.newIntent}
                >
                  <Plus size={17} />
                  开始新对话
                </Link>

                <nav aria-label="产品导航" className="mt-5 space-y-1">
                  {navItems.map((item) => {
                    const Icon = item.icon;
                    return (
                      <NavLink
                        className={({ isActive }) =>
                          cn(
                            "flex h-10 items-center gap-3 rounded-md px-3 text-sm font-medium transition-colors",
                            isActive
                              ? "bg-[#e8f1ed] text-[#315c4d]"
                              : "text-[#625e58] hover:bg-[#f0eeea] hover:text-[#23211f]",
                          )
                        }
                        key={item.to}
                        to={item.to}
                      >
                        <Icon size={17} strokeWidth={1.8} />
                        {item.label}
                        {item.to === "/chat" && totalUnread > 0 ? (
                          <span className="ml-auto inline-flex h-5 min-w-5 items-center justify-center rounded-full bg-[#d95656] px-1 text-[10px] font-semibold text-white">
                            {totalUnread > 99 ? "99+" : totalUnread}
                          </span>
                        ) : null}
                      </NavLink>
                    );
                  })}
                </nav>

                <section className="mt-6 border-t border-[#e4e1dc] pt-5">
                  <h2 className="px-2 text-[11px] font-semibold text-[#817d76]">
                    连接旅程
                  </h2>
                  <ConnectionJourneyNav variant="sidebar" />
                </section>

                <section className="mt-5 min-h-0 border-t border-[#e4e1dc] pt-5">
                  <div className="flex items-center justify-between px-2">
                    <h2 className="text-[11px] font-semibold text-[#817d76]">
                      最近意图
                    </h2>
                    <Link
                      className="text-[11px] font-medium text-[#477062] hover:text-[#315c4d]"
                      to={connectionPaths.intents}
                    >
                      全部
                    </Link>
                  </div>
                  <div className="mt-2 space-y-0.5">
                    {recentIntents.length > 0 ? (
                      recentIntents.map((intent) => (
                        <Link
                          className="group flex min-w-0 items-center gap-2.5 rounded-md px-2 py-2 text-xs text-[#6f6b65] hover:bg-[#f0eeea] hover:text-[#34312e]"
                          key={intent.id}
                          to={connectionIntentPath(intent.id)}
                        >
                          <span
                            className={cn(
                              "h-1.5 w-1.5 shrink-0 rounded-full",
                              intent.status === "active"
                                ? "bg-[#4f846f]"
                                : "bg-[#bbb6ae]",
                            )}
                          />
                          <span className="min-w-0 flex-1 truncate">
                            {intent.title}
                          </span>
                          <Clock3
                            className="shrink-0 text-[#a19c94] opacity-0 transition-opacity group-hover:opacity-100"
                            size={12}
                          />
                        </Link>
                      ))
                    ) : (
                      <p className="px-2 py-3 text-xs leading-5 text-[#9a958d]">
                        暂无最近意图
                      </p>
                    )}
                  </div>
                </section>

                <div className="mt-auto border-t border-[#e4e1dc] pt-4">
                  <div className="flex items-center gap-2.5 px-2">
                    {profile.avatar ? (
                      <AuthenticatedImage
                        alt={profile.nickname || "清华同学"}
                        className="h-8 w-8 shrink-0 rounded-md bg-[#e9f1ed] object-cover"
                        src={profile.avatar}
                      />
                    ) : (
                      <span className="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-md bg-[#e9f1ed] text-sm font-semibold text-[#27614f]">
                        {(profile.nickname || "同学").slice(0, 1)}
                      </span>
                    )}
                    <span className="min-w-0 flex-1">
                      <span className="block truncate text-sm font-medium text-[#302e2b]">
                        {profile.nickname || "清华同学"}
                      </span>
                      <span className="flex items-center gap-1 text-[11px] text-[#77736c]">
                        <ShieldCheck size={11} />
                        身份已验证
                      </span>
                    </span>
                    <button
                      aria-label="分享我的名片"
                      className="inline-flex h-8 w-8 items-center justify-center rounded-md text-[#817d76] transition-colors hover:bg-[#e7efeb] hover:text-[#315c4d]"
                      onClick={() => setShareDrawerOpen(true)}
                      title="分享我的名片"
                      type="button"
                    >
                      <Share2 size={15} />
                    </button>
                    <button
                      aria-label="退出登录"
                      className="inline-flex h-8 w-8 items-center justify-center rounded-md text-[#817d76] transition-colors hover:bg-[#eeeae5] hover:text-[#8b3f3f]"
                      disabled={loggingOut}
                      onClick={() => void handleLogout()}
                      title="退出登录"
                      type="button"
                    >
                      <LogOut size={16} />
                    </button>
                  </div>
                </div>
              </aside>
            </Panel>

            <WorkspaceResizeHandle label="调整主侧栏宽度" />
          </>
        ) : null}

        <Panel
          className="relative min-w-0"
          groupResizeBehavior="preserve-relative-size"
          id="content"
          minSize={isDesktop ? 560 : 0}
        >
          {isDesktop && navigationCollapsed ? (
            <WorkspaceEdgeButton
              label="显示主侧栏"
              onClick={() => navigationPanelRef.current?.expand()}
              side="left"
            />
          ) : null}

          <header className="sticky top-0 z-40 flex h-15 items-center justify-between border-b border-[#dfdcd6] bg-[rgba(251,250,248,0.94)] px-4 backdrop-blur-lg lg:hidden">
            <Brand />
            <div className="flex items-center gap-1">
              <NotificationCenter mobile />
              <Link
                aria-label="发布新意图"
                className="inline-flex h-10 w-10 items-center justify-center rounded-md bg-[#263b35] text-white"
                title="开始新对话"
                to={connectionPaths.newIntent}
              >
                <Plus size={18} />
              </Link>
              <button
                aria-label={mobileMenuOpen ? "关闭菜单" : "打开菜单"}
                className="inline-flex h-10 w-10 items-center justify-center rounded-md text-[#5f5b55] hover:bg-[#efede9]"
                onClick={() => setMobileMenuOpen((open) => !open)}
                type="button"
              >
                {mobileMenuOpen ? <X size={19} /> : <Menu size={19} />}
              </button>
            </div>
          </header>

          {mobileMenuOpen ? (
            <>
              <button
                aria-label="关闭菜单"
                className="fixed inset-0 top-15 z-20 cursor-default bg-black/10 lg:hidden"
                onClick={() => setMobileMenuOpen(false)}
                type="button"
              />
              <div className="fixed inset-x-0 top-15 z-30 border-b border-[#dfdcd6] bg-[#fbfaf8] p-3 shadow-[0_12px_32px_rgba(35,33,31,0.08)] lg:hidden">
              <button
                className="flex h-11 w-full items-center gap-3 rounded-md px-3 text-sm text-[#5f5b55] hover:bg-[#efede9]"
                onClick={() => {
                  setMobileMenuOpen(false);
                  setShareDrawerOpen(true);
                }}
                type="button"
              >
                <Share2 size={17} />
                分享我的名片
              </button>
              <Link
                className="flex h-11 items-center gap-3 rounded-md px-3 text-sm text-[#5f5b55] hover:bg-[#efede9]"
                onClick={() => setMobileMenuOpen(false)}
                to={connectionPaths.intents}
              >
                <ClipboardList size={17} />
                我的意图
              </Link>
              <Link
                className="flex h-11 items-center gap-3 rounded-md px-3 text-sm text-[#5f5b55] hover:bg-[#efede9]"
                onClick={() => setMobileMenuOpen(false)}
                to={connectionPaths.rooms}
              >
                <MessagesSquare size={17} />
                连接房间
              </Link>
              <Link
                className="flex h-11 items-center gap-3 rounded-md px-3 text-sm text-[#5f5b55] hover:bg-[#efede9]"
                onClick={() => setMobileMenuOpen(false)}
                to="/account"
              >
                <UserRound size={17} />
                账户设置
              </Link>
              <button
                className="flex h-11 w-full items-center gap-3 rounded-md px-3 text-sm text-[#8b3f3f] hover:bg-[#f5eaea]"
                disabled={loggingOut}
                onClick={() => void handleLogout()}
                type="button"
              >
                <LogOut size={17} />
                {loggingOut ? "正在退出..." : "退出登录"}
              </button>
              </div>
            </>
          ) : null}

          {inConnectionJourney ? <ConnectionJourneyNav /> : null}

          <main
            className={cn(
              "pb-[calc(4rem_+_env(safe-area-inset-bottom))] lg:h-screen lg:min-h-0 lg:pb-0",
              inConnectionJourney
                ? "min-h-[calc(100dvh_-_6.5rem)]"
                : "min-h-[calc(100dvh_-_3.75rem)]",
            )}
          >
            <Outlet />
          </main>

          <nav
            aria-label="移动端导航"
            className="fixed inset-x-0 bottom-0 z-40 grid h-[calc(4rem_+_env(safe-area-inset-bottom))] grid-cols-5 border-t border-[#dedbd5] bg-[rgba(251,250,248,0.96)] px-1 pb-[env(safe-area-inset-bottom)] backdrop-blur-lg lg:hidden"
          >
            {navItems.map((item) => {
                const Icon = item.icon;
                return (
                  <NavLink
                    className={({ isActive }) =>
                      cn(
                        "relative flex min-w-0 flex-col items-center justify-center gap-1 text-[10px] font-medium",
                        isActive ? "text-[#315c4d]" : "text-[#77736c]",
                      )
                    }
                    key={item.to}
                    onClick={() => setMobileMenuOpen(false)}
                    to={item.to}
                  >
                    <Icon size={18} strokeWidth={1.8} />
                    <span className="truncate">{item.mobileLabel}</span>
                    {item.to === "/chat" && totalUnread > 0 ? (
                      <span className="absolute left-1/2 top-1 ml-1 inline-flex h-4 min-w-4 items-center justify-center rounded-full bg-[#d95656] px-1 text-[9px] font-semibold text-white">
                        {totalUnread > 9 ? "9+" : totalUnread}
                      </span>
                    ) : null}
                  </NavLink>
                );
              })}
          </nav>
        </Panel>
      </Group>
      <ProfileShareDrawer
        onClose={() => setShareDrawerOpen(false)}
        open={shareDrawerOpen}
        userId={session?.userId ?? 0}
      />
    </div>
  );
}
