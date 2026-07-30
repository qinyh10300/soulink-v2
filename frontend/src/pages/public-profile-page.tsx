import { ArrowLeft, Heart, MessageCircle, UserRound } from 'lucide-react'
import { useCallback, useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'

import { getErrorMessage } from '@/lib/api'
import { fetchPlaygroundPosts, type PlaygroundPost } from '@/lib/playground-api'
import { fetchPublicProfile, type PublicProfileDetails } from '@/lib/profile-api'

function formatDate(value: string) {
  return new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  }).format(new Date(value))
}

function avatarLabel(name: string) {
  return name.trim().slice(0, 1)
}

export function PublicProfilePage() {
  const { userId } = useParams()
  const numericUserId = Number(userId)
  const [profile, setProfile] = useState<PublicProfileDetails | null>(null)
  const [posts, setPosts] = useState<PlaygroundPost[]>([])
  const [notice, setNotice] = useState('正在加载个人主页...')

  const loadProfile = useCallback(async () => {
    if (!Number.isFinite(numericUserId) || numericUserId <= 0) {
      setNotice('用户不存在')
      return
    }

    try {
      const [profileData, postData] = await Promise.all([
        fetchPublicProfile(numericUserId),
        fetchPlaygroundPosts(numericUserId),
      ])
      setProfile(profileData)
      setPosts(postData)
      setNotice('')
    } catch (error) {
      setNotice(getErrorMessage(error, '个人主页加载失败，请稍后重试'))
    }
  }, [numericUserId])

  useEffect(() => {
    void loadProfile()
  }, [loadProfile])

  return (
    <div className="min-h-screen bg-[#f7f2ec] px-4 pb-12 pt-6 sm:px-6 lg:px-8">
      <div className="mx-auto max-w-3xl space-y-6">
        <div className="flex items-center gap-3">
          <Link
            className="inline-flex h-9 w-9 items-center justify-center rounded-full border border-[#eee4db] bg-white/80 text-[#5f4d42] transition-colors hover:bg-white"
            to="/playground"
          >
            <ArrowLeft className="h-4 w-4" />
          </Link>
          <h1 className="text-xl font-semibold text-[#2f221b]">个人主页</h1>
        </div>

        {profile ? (
          <section className="overflow-hidden rounded-[32px] border border-[#eee4db] bg-[rgba(255,251,247,0.96)] shadow-[0_18px_54px_rgba(72,50,33,0.06)]">
            <div className="h-32 bg-[radial-gradient(circle_at_top_left,#e2c4a8,#f8efe7_46%,#fffaf5)]" />
            <div className="px-6 pb-6">
              <div className="-mt-12 flex h-24 w-24 items-center justify-center rounded-full border-4 border-[#fffaf5] bg-[#f1e2d4] text-4xl font-semibold text-[#8f5c42] shadow-[0_14px_30px_rgba(72,50,33,0.12)]">
                {avatarLabel(profile.nickname) || <UserRound className="h-8 w-8" />}
              </div>
              <div className="mt-4 space-y-2">
                <h2 className="text-3xl font-semibold text-[#2f221b]">{profile.nickname}</h2>
                <p className="text-sm text-[#8a7467]">
                  {[profile.schoolName, profile.department, profile.grade].filter(Boolean).join(' · ')}
                </p>
                <p className="text-sm text-[#a08d81]">
                  加入于 {formatDate(profile.createdAt)}
                </p>
              </div>
            </div>
          </section>
        ) : null}

        {notice ? (
          <div className="rounded-2xl border border-[#eadfd5] bg-white/75 px-4 py-3 text-sm text-[#7b675b]">
            {notice}
          </div>
        ) : null}

        <section className="space-y-4">
          <h3 className="text-lg font-semibold text-[#2f221b]">TA 的动态</h3>
          {posts.length === 0 && profile ? (
            <div className="rounded-[28px] border border-[#eee4db] bg-white/78 p-6 text-sm text-[#8a7467]">
              暂时还没有公开动态。
            </div>
          ) : null}
          {posts.map((post) => (
            <article
              className="rounded-[28px] border border-[#eee4db] bg-[rgba(255,251,247,0.96)] p-5 shadow-[0_18px_54px_rgba(72,50,33,0.06)]"
              key={post.id}
            >
              <p className="text-xs text-[#a08d81]">{formatDate(post.createdAt)}</p>
              {post.content ? (
                <p className="mt-3 whitespace-pre-wrap break-words text-[15px] leading-8 text-[#3f332d]">
                  {post.content}
                </p>
              ) : null}
              {post.imageUrls.length > 0 ? (
                <div className="mt-4 grid grid-cols-1 gap-3 sm:grid-cols-2">
                  {post.imageUrls.map((imageUrl) => (
                    <img
                      alt=""
                      className="max-h-72 w-full rounded-[22px] border border-[#eee4db] object-cover"
                      key={imageUrl}
                      loading="lazy"
                      src={imageUrl}
                    />
                  ))}
                </div>
              ) : null}
              <div className="mt-4 flex items-center gap-4 text-sm text-[#8a7467]">
                <span className="inline-flex items-center gap-1">
                  <Heart className="h-4 w-4" />
                  {post.likeCount}
                </span>
                <span className="inline-flex items-center gap-1">
                  <MessageCircle className="h-4 w-4" />
                  {post.commentCount}
                </span>
              </div>
            </article>
          ))}
        </section>
      </div>
    </div>
  )
}
