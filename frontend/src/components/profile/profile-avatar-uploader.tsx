import { getAvatarFallbackLabel } from '@/lib/avatar'

interface ProfileAvatarUploaderProps {
  avatar: string
  nickname: string
  onFileChange: (file: File) => void
}

export function ProfileAvatarUploader({
  avatar,
  nickname,
  onFileChange,
}: ProfileAvatarUploaderProps) {
  return (
    <div className="flex items-center gap-5">
      {avatar ? (
        <img
          alt={nickname}
          className="h-20 w-20 rounded-full border border-white/80 object-cover shadow-[0_14px_28px_rgba(73,49,33,0.12)]"
          src={avatar}
        />
      ) : (
        <div className="flex h-20 w-20 items-center justify-center rounded-full bg-[#f3e8de] text-3xl text-[#8f6548]">
          {getAvatarFallbackLabel(nickname)}
        </div>
      )}

      <div className="space-y-2">
        <p className="text-base font-medium text-[#2f221b]">头像</p>
        <label className="inline-flex cursor-pointer items-center rounded-full border border-[#e3d7cc] bg-white px-5 py-2.5 text-sm text-[#6b5a4d] transition-colors hover:border-[#d0beb0] hover:bg-[#fdfaf7]">
          更换头像
          <input
            accept="image/*"
            className="hidden"
            onChange={(event) => {
              const file = event.target.files?.[0]

              if (file) {
                onFileChange(file)
              }
            }}
            type="file"
          />
        </label>
      </div>
    </div>
  )
}
