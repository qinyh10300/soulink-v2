import {
  ArrowRight,
  CalendarCheck2,
  ChevronRight,
  Clock3,
  MessageSquareReply,
  MapPin,
  PanelLeftClose,
  PanelRightClose,
  RotateCcw,
  RefreshCw,
  Send,
  Sparkles,
  UsersRound,
} from "lucide-react";
import { useEffect, useMemo, useState, type KeyboardEvent } from "react";
import { Group, Panel, usePanelRef } from "react-resizable-panels";
import { Link, useNavigate } from "react-router-dom";

import {
  IntentCategoryIcon,
  intentCategoryMeta,
} from "@/components/connection/intent-category";
import { InferenceBadge } from "@/components/connection/inference-badge";
import { WorkspaceEdgeButton } from "@/components/workspace/workspace-edge-button";
import { WorkspaceResizeHandle } from "@/components/workspace/workspace-resize-handle";
import { useMediaQuery } from "@/hooks/use-media-query";
import { usePersistentPanelLayout } from "@/hooks/use-persistent-panel-layout";
import { getErrorMessage } from "@/lib/api";
import {
  fetchIntentFeed,
  fetchDailyRecommendations,
  fetchMatchRooms,
  fetchMyIntents,
  type ConnectionIntent,
  type DailyRecommendations,
  type IntentCategory,
  type MatchRoom,
} from "@/lib/connection-api";
import {
  connectionIntentPath,
  connectionPaths,
  connectionRoomPath,
} from "@/lib/connection-routes";
import { fetchMeetups, type Meetup } from "@/lib/meetup-api";
import { getStoredUserProfile } from "@/lib/user-profile-storage";

const starters: Array<{
  category: IntentCategory;
  label: string;
  prompt: string;
}> = [
  {
    category: "meal",
    label: "找饭搭子",
    prompt: "今晚想在紫荆找个饭搭子，最好也在做 AI 产品",
  },
  {
    category: "study",
    label: "找学习搭子",
    prompt: "这周找一个安静高效的学习搭子，一起准备期末",
  },
  {
    category: "project",
    label: "找项目伙伴",
    prompt: "找一个能一起做产品的项目伙伴，先用两周做出原型",
  },
  {
    category: "cofounder",
    label: "找 CTO / 联创",
    prompt: "为我的 AI 项目找 CTO 或联合创始人，希望能长期投入",
  },
];

function formatCampusDate() {
  return new Intl.DateTimeFormat("zh-CN", {
    month: "long",
    day: "numeric",
    weekday: "short",
  }).format(new Date());
}

function formatMeetingTime(value?: string) {
  if (!value) {
    return "待确认时间";
  }
  return new Intl.DateTimeFormat("zh-CN", {
    month: "numeric",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
    hour12: false,
  }).format(new Date(value));
}

function formatRefreshTime(value?: string) {
  if (!value) {
    return "";
  }
  return new Intl.DateTimeFormat("zh-CN", {
    hour: "2-digit",
    minute: "2-digit",
    hour12: false,
  }).format(new Date(value));
}

function meetupPath(meetup: Meetup) {
  if (meetup.context.type === "match_room") {
    return connectionRoomPath(meetup.context.id);
  }
  if (meetup.context.type === "group") {
    return `/chat/group/${meetup.context.id}`;
  }
  return `/chat/direct/${meetup.context.id}`;
}

function DailyRecommendationList({ daily }: { daily: DailyRecommendations }) {
  if (daily.items.length === 0) {
    return null;
  }

  return (
    <section className="border-t border-[#dedad4] pt-6 text-left">
      <div className="flex items-center justify-between gap-3">
        <div>
          <h3 className="text-sm font-semibold text-[#34312e]">每日精选</h3>
          <p className="mt-0.5 text-[11px] text-[#817d76]">今天只看这 3 个连接机会</p>
        </div>
        {daily.refreshAt ? (
          <span className="text-[10px] text-[#8a867f]">
            {formatRefreshTime(daily.refreshAt)} 更新
          </span>
        ) : null}
      </div>
      <div className="mt-3 grid gap-2 sm:grid-cols-3">
        {daily.items.map((recommendation, index) => (
          <article
            className="min-w-0 rounded-md border border-[#dedad4] bg-[#fbfaf8] p-3 transition-colors hover:border-[#aebdb7] hover:bg-white"
            key={recommendation.intent.id}
          >
            <div className="flex items-start gap-2">
              <span className="inline-flex h-5 w-5 shrink-0 items-center justify-center rounded-md bg-[#e9efec] text-[10px] font-semibold text-[#46675b]">
                {index + 1}
              </span>
              <Link
                className="min-w-0 flex-1 truncate text-xs font-semibold text-[#45413d] hover:text-[#315c4d]"
                to={connectionIntentPath(recommendation.intent.id)}
              >
                {recommendation.intent.title}
              </Link>
              <span className="shrink-0 text-[10px] font-semibold text-[#5b3477]">
                {recommendation.score}%
              </span>
            </div>
            <p className="mt-2 line-clamp-2 text-[11px] leading-4 text-[#817d76]">
              {recommendation.fitReason}
            </p>
            <InferenceBadge
              className="mt-2"
              compact
              editTo="/profile"
              inference={recommendation.inference}
            />
          </article>
        ))}
      </div>
    </section>
  );
}

export function ConnectionHomePage() {
  const navigate = useNavigate();
  const profile = useMemo(() => getStoredUserProfile(), []);
  const [query, setQuery] = useState("");
  const [feed, setFeed] = useState<ConnectionIntent[]>([]);
  const [myIntents, setMyIntents] = useState<ConnectionIntent[]>([]);
  const [rooms, setRooms] = useState<MatchRoom[]>([]);
  const [meetups, setMeetups] = useState<Meetup[]>([]);
  const [dailyRecommendations, setDailyRecommendations] =
    useState<DailyRecommendations | null>(null);
  const [dailyUnavailable, setDailyUnavailable] = useState(false);
  const [reloadKey, setReloadKey] = useState(0);
  const [loading, setLoading] = useState(true);
  const [notice, setNotice] = useState("");
  const isWideWorkspace = useMediaQuery("(min-width: 1280px)");
  const assistantPanelRef = usePanelRef();
  const contextPanelRef = usePanelRef();
  const homeLayout = usePersistentPanelLayout(
    "soulink-home-workbench-pixels-v2",
    {
      assistant: {},
      context: {
        defaultSize: 320,
        minSize: 280,
        rememberSize: true,
      },
    },
  );
  const [assistantCollapsed, setAssistantCollapsed] = useState(() =>
    homeLayout.isInitiallyCollapsed("assistant"),
  );
  const [contextCollapsed, setContextCollapsed] = useState(() =>
    homeLayout.isInitiallyCollapsed("context"),
  );

  useEffect(() => {
    let cancelled = false;

    const load = async () => {
      setLoading(true);
      setNotice("");
      setDailyUnavailable(false);
      const now = new Date();
      const meetupEnd = new Date(now.getTime() + 30 * 24 * 60 * 60 * 1000);
      const [feedResult, intentsResult, roomsResult, dailyResult, meetupResult] =
        await Promise.allSettled([
          fetchIntentFeed(),
          fetchMyIntents(),
          fetchMatchRooms(),
          fetchDailyRecommendations(),
          fetchMeetups(undefined, now, meetupEnd),
        ]);

      if (cancelled) {
        return;
      }

      const nextFeed =
        feedResult.status === "fulfilled" ? feedResult.value : [];
      const nextIntents =
        intentsResult.status === "fulfilled" ? intentsResult.value : [];
      const nextRooms =
        roomsResult.status === "fulfilled" ? roomsResult.value : [];

      setFeed(nextFeed);
      setMyIntents(nextIntents);
      setRooms(nextRooms);
      setMeetups(
        meetupResult.status === "fulfilled" ? meetupResult.value : [],
      );
      setDailyRecommendations(
        dailyResult.status === "fulfilled" ? dailyResult.value : null,
      );
      setDailyUnavailable(dailyResult.status === "rejected");

      if (
        feedResult.status === "rejected" &&
        intentsResult.status === "rejected" &&
        roomsResult.status === "rejected" &&
        dailyResult.status === "rejected" &&
        meetupResult.status === "rejected"
      ) {
        setNotice(getErrorMessage(feedResult.reason, "首页暂时无法加载"));
      }
      setLoading(false);
    };

    void load();
    return () => {
      cancelled = true;
    };
  }, [reloadKey]);

  const startConversation = (value = query) => {
    const text = value.trim();
    navigate(connectionPaths.newIntent, {
      state: text ? { text } : undefined,
    });
  };

  const handleKeyDown = (event: KeyboardEvent<HTMLTextAreaElement>) => {
    if (
      event.key === "Enter" &&
      !event.shiftKey &&
      !event.nativeEvent.isComposing
    ) {
      event.preventDefault();
      if (query.trim()) {
        startConversation();
      }
    }
  };

  const activeIntents = myIntents.filter(
    (intent) => intent.status === "active",
  );
  const pendingMeetups = meetups
    .filter((meetup) =>
      ["proposed", "confirmed", "reschedule_requested"].includes(
        meetup.status,
      ),
    )
    .sort(
      (left, right) =>
        new Date(left.startsAt).getTime() - new Date(right.startsAt).getTime(),
    );
  const upcomingRoom = rooms
    .filter((room) => room.status === "open")
    .sort((left, right) => {
      const leftTime = left.scheduledAt
        ? new Date(left.scheduledAt).getTime()
        : Number.MAX_SAFE_INTEGER;
      const rightTime = right.scheduledAt
        ? new Date(right.scheduledAt).getTime()
        : Number.MAX_SAFE_INTEGER;
      return leftTime - rightTime;
    })[0];
  const responseMeetup = pendingMeetups.find((meetup) => meetup.canRespond);
  const intentWithResponses = [...activeIntents]
    .filter((intent) => intent.interestCount > 0)
    .sort((left, right) => right.interestCount - left.interestCount)[0];
  const roomNeedsPlan = rooms.find(
    (room) => room.status === "open" && (!room.scheduledAt || !room.selectedPlace),
  );
  const currentTask = responseMeetup
    ? {
        kind: "meetup" as const,
        eyebrow: "等待你的回应",
        title: responseMeetup.title,
        description: responseMeetup.reminder.text,
        action: "查看约见",
        to: meetupPath(responseMeetup),
      }
    : intentWithResponses
      ? {
          kind: "response" as const,
          eyebrow: `${intentWithResponses.interestCount} 个新回应`,
          title: intentWithResponses.title,
          description: "先看附言和相关资料，再决定是否愿意认识。",
          action: "处理回应",
          to: connectionIntentPath(intentWithResponses.id),
        }
      : roomNeedsPlan
        ? {
            kind: "room" as const,
            eyebrow: "双方已经同意",
            title: `和 ${roomNeedsPlan.other.nickname} 确认见面安排`,
            description: roomNeedsPlan.intentTitle,
            action: "进入房间",
            to: connectionRoomPath(roomNeedsPlan.id),
          }
        : activeIntents[0]
          ? {
              kind: "intent" as const,
              eyebrow: "正在寻找",
              title: activeIntents[0].title,
              description: "查看本轮推荐，或补充更具体的时间与地点。",
              action: "查看进展",
              to: connectionIntentPath(activeIntents[0].id),
            }
          : {
              kind: "start" as const,
              eyebrow: "开始一次具体连接",
              title: "今天想认识谁？",
              description: "从一起吃饭、学习、运动或做项目开始。",
              action: "开始描述",
              to: connectionPaths.newIntent,
            };

  const resetHomeLayout = () => {
    homeLayout.reset();
    assistantPanelRef.current?.expand();
    contextPanelRef.current?.expand();
    contextPanelRef.current?.resize(320);
  };

  const collapseAssistant = () => {
    if (contextCollapsed) {
      contextPanelRef.current?.expand();
    }
    assistantPanelRef.current?.collapse();
  };

  const collapseContext = () => {
    if (assistantCollapsed) {
      assistantPanelRef.current?.expand();
    }
    contextPanelRef.current?.collapse();
  };

  const assistantPanel = (
    <section className="flex h-full min-h-0 min-w-0 flex-col">
      <header className="flex h-16 shrink-0 items-center justify-between border-b border-[#dfdcd6] px-4 sm:px-6">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-sm font-semibold text-[#292724]">
              认识
            </h1>
            <span className="h-1.5 w-1.5 rounded-full bg-[#4f846f]" />
          </div>
          <p className="mt-0.5 text-[11px] text-[#817d76]">
            {formatCampusDate()} · 清华园
          </p>
        </div>

        <div className="flex items-center gap-3 text-[11px] text-[#77736c] xl:hidden">
          <Link className="hover:text-[#315c4d]" to={connectionPaths.intents}>
            {activeIntents.length} 个意图
          </Link>
          <span className="h-3 w-px bg-[#d6d2cb]" />
          <Link className="hover:text-[#315c4d]" to={connectionPaths.rooms}>
            {rooms.filter((room) => room.status === "open").length} 个连接
          </Link>
        </div>

        {isWideWorkspace ? (
          <div className="flex items-center gap-0.5">
            <button
              aria-label="恢复首页默认布局"
              className="inline-flex h-8 w-8 items-center justify-center rounded-md text-[#8a857e] transition-colors hover:bg-[#efede9] hover:text-[#445f55] focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-[#527b6d]"
              onClick={resetHomeLayout}
              title="恢复首页默认布局"
              type="button"
            >
              <RotateCcw size={14} />
            </button>
            <button
              aria-label="隐藏 SouLink 助手"
              className="inline-flex h-8 w-8 items-center justify-center rounded-md text-[#8a857e] transition-colors hover:bg-[#efede9] hover:text-[#445f55] focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-[#527b6d]"
              onClick={collapseAssistant}
              title="隐藏 SouLink 助手"
              type="button"
            >
              <PanelLeftClose size={15} />
            </button>
          </div>
        ) : null}
      </header>

      <div className="min-h-0 flex-1 overflow-y-auto">
        <div className="mx-auto w-full max-w-4xl px-5 py-6 sm:px-8 sm:py-8">
          <div className="flex items-end justify-between gap-4">
            <div>
              <p className="text-xs font-medium text-[#527164]">
                {formatCampusDate()} · 清华园
              </p>
              <h2 className="mt-1 text-2xl font-semibold leading-tight text-[#262421]">
                {profile.nickname ? `${profile.nickname}，今天先做这一件事` : "今天先做这一件事"}
              </h2>
            </div>
            <Link
              aria-label="开始新的寻找"
              className="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-md border border-[#cfd7d3] text-[#315c4d] hover:bg-[#eaf1ed]"
              title="开始新的寻找"
              to={connectionPaths.newIntent}
            >
              <Sparkles size={16} />
            </Link>
          </div>

          {loading ? (
            <div className="mt-5 flex min-h-32 items-center justify-center border-y border-[#dedad4] text-sm text-[#77736c]">
              正在整理今天的连接进度...
            </div>
          ) : (
            <Link
              className="group mt-5 flex min-h-32 items-start gap-4 border-y border-[#ccd5d1] bg-[#f0f5f2] px-4 py-5 transition-colors hover:bg-[#eaf2ee] sm:px-5"
              to={currentTask.to}
            >
              <span className="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-md bg-[#315c4d] text-white">
                {currentTask.kind === "meetup" ? (
                  <CalendarCheck2 size={18} />
                ) : currentTask.kind === "response" ? (
                  <MessageSquareReply size={18} />
                ) : (
                  <Sparkles size={18} />
                )}
              </span>
              <span className="min-w-0 flex-1">
                <span className="text-[11px] font-semibold text-[#527164]">
                  {currentTask.eyebrow}
                </span>
                <span className="mt-1 block text-base font-semibold leading-6 text-[#303b36]">
                  {currentTask.title}
                </span>
                <span className="mt-1 block text-xs leading-5 text-[#67736d]">
                  {currentTask.description}
                </span>
              </span>
              <span className="mt-1 inline-flex shrink-0 items-center gap-1 text-xs font-semibold text-[#315c4d]">
                {currentTask.action}
                <ArrowRight
                  className="transition-transform group-hover:translate-x-0.5"
                  size={14}
                />
              </span>
            </Link>
          )}

          {dailyRecommendations ? (
            <div className="mt-7">
              <DailyRecommendationList daily={dailyRecommendations} />
            </div>
          ) : dailyUnavailable && !loading ? (
            <div className="mt-7 flex items-center justify-between gap-3 border-t border-[#dedad4] py-5">
              <p className="text-xs text-[#80665c]">每日精选暂时没有同步</p>
              <button
                aria-label="重新加载每日精选"
                className="inline-flex h-8 w-8 items-center justify-center rounded-md text-[#527164] hover:bg-[#eaf1ed]"
                onClick={() => setReloadKey((value) => value + 1)}
                title="重新加载"
                type="button"
              >
                <RefreshCw size={14} />
              </button>
            </div>
          ) : null}

          {pendingMeetups.length > 0 ? (
            <section className="mt-7 border-t border-[#dedad4] pt-6">
              <div className="flex items-center justify-between gap-3">
                <div>
                  <h3 className="text-sm font-semibold text-[#34312e]">待确认约见</h3>
                  <p className="mt-0.5 text-[11px] text-[#817d76]">
                    {pendingMeetups.length} 个安排仍在进行中
                  </p>
                </div>
                <Link
                  className="text-xs font-semibold text-[#477062]"
                  to="/calendar"
                >
                  时间表
                </Link>
              </div>
              <div className="mt-3 divide-y divide-[#e4e1dc] border-y border-[#e4e1dc]">
                {pendingMeetups.slice(0, 3).map((meetup) => (
                  <Link
                    className="flex items-center gap-3 py-3.5"
                    key={meetup.id}
                    to={meetupPath(meetup)}
                  >
                    <CalendarCheck2 className="shrink-0 text-[#5b3477]" size={16} />
                    <span className="min-w-0 flex-1">
                      <span className="block truncate text-xs font-semibold text-[#45413d]">
                        {meetup.title}
                      </span>
                      <span className="mt-0.5 block truncate text-[11px] text-[#817d76]">
                        {formatMeetingTime(meetup.startsAt)} · {meetup.location}
                      </span>
                    </span>
                    {meetup.canRespond ? (
                      <span className="shrink-0 rounded-md bg-[#f1eaf4] px-2 py-1 text-[10px] font-semibold text-[#5b3477]">
                        待回应
                      </span>
                    ) : (
                      <ChevronRight className="shrink-0 text-[#a19c94]" size={14} />
                    )}
                  </Link>
                ))}
              </div>
            </section>
          ) : null}

          <section className="mt-7 border-t border-[#dedad4] pt-6">
            <h3 className="text-sm font-semibold text-[#34312e]">开始新的寻找</h3>
            <div className="mt-3 grid gap-2 sm:grid-cols-2">
              {starters.map(({ category, label, prompt }) => {
                const meta = intentCategoryMeta[category];
                const Icon = meta.icon;
                return (
                  <button
                    className="group flex min-h-14 items-center gap-3 rounded-md border border-[#dedad4] bg-[#fbfaf8] px-3 py-2.5 text-left transition-colors hover:border-[#9eaea8] hover:bg-white"
                    key={category}
                    onClick={() => startConversation(prompt)}
                    type="button"
                  >
                    <span
                      className={`inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-md ${meta.surfaceClassName}`}
                    >
                      <Icon className={meta.iconClassName} size={15} />
                    </span>
                    <span className="min-w-0 flex-1">
                      <span className="block text-xs font-semibold text-[#45413d]">
                        {label}
                      </span>
                      <span className="mt-0.5 block truncate text-[11px] text-[#817d76]">
                        {prompt}
                      </span>
                    </span>
                    <ArrowRight
                      className="shrink-0 text-[#a19c94] transition-transform group-hover:translate-x-0.5 group-hover:text-[#477062]"
                      size={15}
                    />
                  </button>
                );
              })}
            </div>
          </section>

          {notice ? <p className="mt-5 text-xs text-[#8c463d]">{notice}</p> : null}
        </div>
      </div>

      <footer className="shrink-0 border-t border-[#dfdcd6] bg-[rgba(247,246,243,0.96)] px-3 py-3 backdrop-blur-lg sm:px-6 sm:py-4">
        <div className="mx-auto max-w-3xl">
          <div className="flex items-end gap-2 rounded-lg border border-[#d6d2cb] bg-white p-2 shadow-[0_8px_24px_rgba(35,33,31,0.06)] focus-within:border-[#6f9185] focus-within:ring-3 focus-within:ring-[#55786c]/10">
            <textarea
              aria-label="描述你想认识的人"
              autoFocus
              className="max-h-28 min-h-10 min-w-0 flex-1 resize-none border-0 bg-transparent px-2 py-2 text-sm leading-6 text-[#292724] outline-none placeholder:text-[#9a958d]"
              maxLength={1200}
              onChange={(event) => setQuery(event.target.value)}
              onKeyDown={handleKeyDown}
              placeholder="描述你想找的人，以及想一起做什么"
              rows={1}
              value={query}
            />
            <button
              aria-label="发送"
              className="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-md bg-[#263b35] text-white transition-colors hover:bg-[#1d302a] disabled:cursor-not-allowed disabled:bg-[#cbc8c2]"
              disabled={!query.trim()}
              onClick={() => startConversation()}
              title="发送"
              type="button"
            >
              <Send size={17} />
            </button>
          </div>
        </div>
      </footer>
    </section>
  );

  const contextPanel = (
    <aside className="h-full min-h-0 overflow-y-auto bg-[#fbfaf8]">
      <header className="flex h-16 items-center justify-between border-b border-[#dfdcd6] px-5">
        <h2 className="text-sm font-semibold text-[#302e2b]">你的上下文</h2>
        <div className="flex items-center gap-1.5">
          <span className="text-[11px] text-[#817d76]">
            {loading ? "同步中" : "刚刚同步"}
          </span>
          <button
            aria-label="隐藏上下文"
            className="inline-flex h-8 w-8 items-center justify-center rounded-md text-[#8a857e] transition-colors hover:bg-[#efede9] hover:text-[#445f55] focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-[#527b6d]"
            onClick={collapseContext}
            title="隐藏上下文"
            type="button"
          >
            <PanelRightClose size={15} />
          </button>
        </div>
      </header>

      <section className="border-b border-[#e2dfd9] px-5 py-5">
        <div className="flex items-center justify-between gap-3">
          <h3 className="text-xs font-semibold text-[#5f5b55]">进行中的意图</h3>
          <Link
            className="text-[11px] font-medium text-[#477062]"
            to={connectionPaths.intents}
          >
            管理
          </Link>
        </div>
        {activeIntents.length > 0 ? (
          <div className="mt-3">
            {activeIntents.slice(0, 2).map((intent) => (
              <Link
                className="flex items-center gap-3 border-b border-[#ece9e4] py-3 last:border-b-0"
                key={intent.id}
                to={connectionIntentPath(intent.id)}
              >
                <IntentCategoryIcon category={intent.category} size={14} />
                <span className="min-w-0 flex-1">
                  <span className="block truncate text-xs font-semibold text-[#45413d]">
                    {intent.title}
                  </span>
                  <span className="mt-0.5 block text-[11px] text-[#817d76]">
                    {intent.interestCount} 个回应
                  </span>
                </span>
                <ChevronRight className="text-[#a19c94]" size={14} />
              </Link>
            ))}
          </div>
        ) : (
          <p className="mt-3 text-xs leading-5 text-[#817d76]">暂无活跃意图</p>
        )}
      </section>

      <section className="border-b border-[#e2dfd9] px-5 py-5">
        <div className="flex items-center justify-between gap-3">
          <h3 className="text-xs font-semibold text-[#5f5b55]">下一次连接</h3>
          <Link
            className="text-[11px] font-medium text-[#477062]"
            to={connectionPaths.rooms}
          >
            全部
          </Link>
        </div>
        {upcomingRoom ? (
          <Link className="mt-3 block" to={connectionRoomPath(upcomingRoom.id)}>
            <div className="flex items-center gap-3">
              <span className="inline-flex h-9 w-9 items-center justify-center rounded-md bg-[#e8f1ed] text-[#356653]">
                <CalendarCheck2 size={17} />
              </span>
              <span className="min-w-0">
                <span className="block truncate text-xs font-semibold text-[#45413d]">
                  和 {upcomingRoom.other.nickname}
                </span>
                <span className="mt-0.5 block truncate text-[11px] text-[#817d76]">
                  {upcomingRoom.intentTitle}
                </span>
              </span>
            </div>
            <div className="mt-3 space-y-1.5 text-[11px] text-[#706c65]">
              <p className="flex items-center gap-2">
                <Clock3 size={13} />
                {formatMeetingTime(upcomingRoom.scheduledAt)}
              </p>
              <p className="flex items-center gap-2">
                <MapPin size={13} />
                {upcomingRoom.selectedPlace || "待确认地点"}
              </p>
            </div>
          </Link>
        ) : (
          <p className="mt-3 text-xs leading-5 text-[#817d76]">
            暂无待安排的连接
          </p>
        )}
      </section>

      <section className="px-5 py-5">
        <div className="flex items-center justify-between gap-3">
          <h3 className="text-xs font-semibold text-[#5f5b55]">校园此刻</h3>
          <Link
            className="text-[11px] font-medium text-[#477062]"
            to={connectionPaths.discover}
          >
            发现
          </Link>
        </div>
        {feed.length > 0 ? (
          <div className="mt-3">
            {feed.slice(0, 3).map((intent) => (
              <Link
                className="flex gap-2.5 border-b border-[#ece9e4] py-2.5 last:border-b-0"
                key={intent.id}
                to={connectionIntentPath(intent.id)}
              >
                <UsersRound
                  className="mt-0.5 shrink-0 text-[#817d76]"
                  size={13}
                />
                <span className="line-clamp-2 text-xs leading-5 text-[#5f5b55]">
                  {intent.title}
                </span>
              </Link>
            ))}
          </div>
        ) : (
          <p className="mt-3 text-xs leading-5 text-[#817d76]">
            暂时没有公开意图
          </p>
        )}
      </section>
    </aside>
  );

  return (
    <div className="h-[calc(100dvh_-_10.5rem_-_env(safe-area-inset-bottom))] overflow-hidden bg-[#f7f6f3] lg:h-screen">
      {isWideWorkspace ? (
        <Group
          className="h-full w-full"
          elementRef={homeLayout.groupElementRef}
          id="soulink-home-workbench"
          onLayoutChanged={homeLayout.onLayoutChanged}
          orientation="horizontal"
          resizeTargetMinimumSize={{ coarse: 24, fine: 11 }}
        >
          <Panel
            className="relative min-w-0"
            collapsedSize={0}
            collapsible
            defaultSize={homeLayout.getDefaultSize("assistant")}
            groupResizeBehavior="preserve-relative-size"
            id="assistant"
            minSize={520}
            onResize={({ inPixels }) => setAssistantCollapsed(inPixels < 1)}
            panelRef={assistantPanelRef}
          >
            {assistantPanel}
            {contextCollapsed ? (
              <WorkspaceEdgeButton
                label="显示上下文"
                onClick={() => contextPanelRef.current?.expand()}
                side="right"
              />
            ) : null}
          </Panel>

          <WorkspaceResizeHandle label="调整助手与上下文宽度" />

          <Panel
            className="relative min-w-0"
            collapsedSize={0}
            collapsible
            defaultSize={homeLayout.getDefaultSize("context")}
            groupResizeBehavior="preserve-pixel-size"
            id="context"
            minSize={280}
            onResize={({ inPixels }) => setContextCollapsed(inPixels < 1)}
            panelRef={contextPanelRef}
          >
            {contextPanel}
            {assistantCollapsed ? (
              <WorkspaceEdgeButton
                label="显示 SouLink 助手"
                onClick={() => assistantPanelRef.current?.expand()}
                side="left"
              />
            ) : null}
          </Panel>
        </Group>
      ) : (
        assistantPanel
      )}
    </div>
  );
}
