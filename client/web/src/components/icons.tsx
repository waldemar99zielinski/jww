import type { SVGProps } from 'react'

type IconProps = SVGProps<SVGSVGElement>

function Icon({ children, ...props }: IconProps) {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth={1.75}
      strokeLinecap="round"
      strokeLinejoin="round"
      {...props}
    >
      {children}
    </svg>
  )
}

export function IconHome(props: IconProps) {
  return (
    <Icon {...props}>
      <path d="M4 11.5 12 4l8 7.5" />
      <path d="M6 10v9h12v-9" />
      <path d="M10 19v-5h4v5" />
    </Icon>
  )
}

export function IconList(props: IconProps) {
  return (
    <Icon {...props}>
      <path d="M7 4h10l2 3v13H5V7Z" />
      <path d="M9 11h6M9 15h6" />
    </Icon>
  )
}

export function IconPlus(props: IconProps) {
  return (
    <Icon {...props} strokeWidth={2}>
      <path d="M12 5v14M5 12h14" />
    </Icon>
  )
}

export function IconTarget(props: IconProps) {
  return (
    <Icon {...props}>
      <circle cx="12" cy="12" r="8" />
      <circle cx="12" cy="12" r="4" />
      <path d="M12 2v4M12 18v4" />
    </Icon>
  )
}

export function IconGear(props: IconProps) {
  return (
    <Icon {...props} strokeWidth={1.6}>
      <circle cx="12" cy="12" r="3" />
      <path d="M19.4 13.5a1.8 1.8 0 0 0 .36 1.98l.07.07a2.17 2.17 0 1 1-3.07 3.07l-.07-.07a1.8 1.8 0 0 0-1.98-.36 1.8 1.8 0 0 0-1.1 1.65V20a2.17 2.17 0 1 1-4.33 0v-.1a1.8 1.8 0 0 0-1.17-1.65 1.8 1.8 0 0 0-1.98.36l-.07.07a2.17 2.17 0 1 1-3.07-3.07l.07-.07a1.8 1.8 0 0 0 .36-1.98 1.8 1.8 0 0 0-1.65-1.1H2a2.17 2.17 0 1 1 0-4.33h.1a1.8 1.8 0 0 0 1.65-1.17 1.8 1.8 0 0 0-.36-1.98l-.07-.07A2.17 2.17 0 1 1 6.4 3.07l.07.07a1.8 1.8 0 0 0 1.98.36H8.6a1.8 1.8 0 0 0 1.1-1.65V2a2.17 2.17 0 1 1 4.33 0v.1a1.8 1.8 0 0 0 1.1 1.65 1.8 1.8 0 0 0 1.98-.36l.07-.07a2.17 2.17 0 1 1 3.07 3.07l-.07.07a1.8 1.8 0 0 0-.36 1.98v.14a1.8 1.8 0 0 0 1.65 1.1H22a2.17 2.17 0 1 1 0 4.33h-.1a1.8 1.8 0 0 0-1.65 1.1Z" />
    </Icon>
  )
}

export function IconBasket(props: IconProps) {
  return (
    <Icon {...props}>
      <path d="M5 9h14l-1.4 9.5a2 2 0 0 1-2 1.5H8.4a2 2 0 0 1-2-1.5L5 9Z" />
      <path d="M8 9V7a4 4 0 0 1 8 0v2M9 13v4M15 13v4" />
    </Icon>
  )
}

export function IconCar(props: IconProps) {
  return (
    <Icon {...props}>
      <path d="M4 16V11l2-5h12l2 5v5" />
      <path d="M4 16h16M7 16v2M17 16v2" />
      <circle cx="7.5" cy="16" r="1.4" />
      <circle cx="16.5" cy="16" r="1.4" />
    </Icon>
  )
}

export function IconFilm(props: IconProps) {
  return (
    <Icon {...props}>
      <rect x="3.5" y="5" width="17" height="14" rx="2" />
      <path d="M8 5v14M16 5v14M3.5 9.5h4M16.5 9.5h4M3.5 14.5h4M16.5 14.5h4" />
    </Icon>
  )
}

export function IconBuilding(props: IconProps) {
  return (
    <Icon {...props}>
      <rect x="5" y="3.5" width="14" height="17" rx="1.5" />
      <path d="M9 7.5h.01M15 7.5h.01M9 11.5h.01M15 11.5h.01M9 15.5h.01M15 15.5h.01" />
      <path d="M10 20.5v-4h4v4" />
    </Icon>
  )
}

export function IconHeart(props: IconProps) {
  return (
    <Icon {...props}>
      <path d="M12 20s-7.5-4.6-9.7-9A5.3 5.3 0 0 1 12 6.3 5.3 5.3 0 0 1 21.7 11c-2.2 4.4-9.7 9-9.7 9Z" />
    </Icon>
  )
}

export function IconChevronDown(props: IconProps) {
  return (
    <Icon {...props} strokeWidth={2}>
      <path d="m6 9 6 6 6-6" />
    </Icon>
  )
}

export function IconChevronRight(props: IconProps) {
  return (
    <Icon {...props} strokeWidth={2}>
      <path d="m9 6 6 6-6 6" />
    </Icon>
  )
}

export function IconClose(props: IconProps) {
  return (
    <Icon {...props} strokeWidth={2}>
      <path d="M6 6l12 12M18 6 6 18" />
    </Icon>
  )
}

export function IconSearch(props: IconProps) {
  return (
    <Icon {...props} strokeWidth={1.9}>
      <circle cx="11" cy="11" r="7" />
      <path d="m21 21-4.5-4.5" />
    </Icon>
  )
}

export function IconCalendar(props: IconProps) {
  return (
    <Icon {...props}>
      <rect x="3.5" y="5.5" width="17" height="15" rx="2.5" />
      <path d="M3.5 10h17M8 3.5v4M16 3.5v4" />
    </Icon>
  )
}

export function IconNote(props: IconProps) {
  return (
    <Icon {...props}>
      <path d="M14.5 4H7a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V8.5L14.5 4Z" />
      <path d="M14 4v4.5h4.5M9 13h6M9 16.5h4" />
    </Icon>
  )
}

export function IconBell(props: IconProps) {
  return (
    <Icon {...props}>
      <path d="M6 10a6 6 0 1 1 12 0c0 4 1.5 5.5 1.5 5.5h-15S6 14 6 10Z" />
      <path d="M10 19a2 2 0 0 0 4 0" />
    </Icon>
  )
}

export function IconDownload(props: IconProps) {
  return (
    <Icon {...props}>
      <path d="M12 4v11m0 0 4-4m-4 4-4-4" />
      <path d="M5 18.5h14" />
    </Icon>
  )
}

export function IconInstall(props: IconProps) {
  return (
    <Icon {...props}>
      <path d="M7 3.5h10a1.5 1.5 0 0 1 1.5 1.5v14a1.5 1.5 0 0 1-1.5 1.5H7A1.5 1.5 0 0 1 5.5 19V5A1.5 1.5 0 0 1 7 3.5Z" />
      <path d="M12 8v5.5" />
      <path d="m9.7 11.2 2.3 2.3 2.3-2.3" />
    </Icon>
  )
}

export function IconInfo(props: IconProps) {
  return (
    <Icon {...props}>
      <circle cx="12" cy="12" r="8.5" />
      <path d="M12 11v5.5" />
      <circle cx="12" cy="8" r="0.6" fill="currentColor" stroke="none" />
    </Icon>
  )
}

export function IconChart(props: IconProps) {
  return (
    <Icon {...props}>
      <path d="M4 20V10M10 20V4M16 20v-7M22 20H2" />
    </Icon>
  )
}

export function IconWallet(props: IconProps) {
  return (
    <Icon {...props}>
      <rect x="3" y="6.5" width="18" height="13" rx="3" />
      <path d="M15.5 11H19a2 2 0 0 1 0 4h-3.5a2 2 0 0 1 0-4Z" />
      <circle cx="18" cy="13" r="0.5" fill="currentColor" stroke="none" />
    </Icon>
  )
}

export function IconTag(props: IconProps) {
  return (
    <Icon {...props}>
      <path d="M12.5 3h6.5v6.5L8.5 20 2 13.5 12.5 3Z" />
      <circle cx="14.5" cy="8.5" r="1.1" fill="currentColor" stroke="none" />
    </Icon>
  )
}

export function IconCamera(props: IconProps) {
  return (
    <Icon {...props}>
      <path d="M4 7.5h3l1.5-2.5h7L17 7.5h3a1 1 0 0 1 1 1v9a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1v-9a1 1 0 0 1 1-1Z" />
      <circle cx="12" cy="12.5" r="3.2" />
    </Icon>
  )
}

export function IconPaperclip(props: IconProps) {
  return (
    <Icon {...props}>
      <path d="M21.4 11.05 12.3 20.15a5.5 5.5 0 0 1-7.78-7.78l9.19-9.19a3.5 3.5 0 1 1 4.95 4.95l-9.2 9.2a1.5 1.5 0 0 1-2.12-2.13l8.49-8.48" />
    </Icon>
  )
}

export function IconImage(props: IconProps) {
  return (
    <Icon {...props}>
      <rect x="3.5" y="5" width="17" height="14" rx="2.5" />
      <circle cx="9" cy="10" r="1.5" />
      <path d="m4.5 17 4.5-4 4 3 2.5-2 4 3.5" />
    </Icon>
  )
}

export function IconRepeat(props: IconProps) {
  return (
    <Icon {...props}>
      <path d="m17 3 3 3-3 3" />
      <path d="M20 6H8a4 4 0 0 0-4 4v1" />
      <path d="m7 21-3-3 3-3" />
      <path d="M4 18h12a4 4 0 0 0 4-4v-1" />
    </Icon>
  )
}

export function IconShirt(props: IconProps) {
  return (
    <Icon {...props}>
      <path d="M9 4 4 7l2.5 3.5L9 9V20h6V9l2.5 1.5L20 7l-5-3q-1.5 2-3 2t-3-2Z" />
    </Icon>
  )
}

export function IconPlane(props: IconProps) {
  return (
    <Icon {...props}>
      <path d="M21 3 3 10.5l7.5 3L14 21 21 3Z" />
      <path d="M10.5 13.5 21 3" />
    </Icon>
  )
}

export function IconBolt(props: IconProps) {
  return (
    <Icon {...props}>
      <path d="M13 2 4 14h6l-1 8 9-12h-6l1-8Z" />
    </Icon>
  )
}

export function IconGraduationCap(props: IconProps) {
  return (
    <Icon {...props}>
      <path d="M2 9 12 4l10 5-10 5L2 9Z" />
      <path d="M6 11.5V16c0 1.5 2.7 3 6 3s6-1.5 6-3v-4.5" />
      <path d="M21 9v5" />
    </Icon>
  )
}

export function IconSparkle(props: IconProps) {
  return (
    <Icon {...props}>
      <path d="M12 3c.5 3 2 5.5 5 6-3 .5-4.5 3-5 6-.5-3-2-5.5-5-6 3-.5 4.5-3 5-6Z" />
    </Icon>
  )
}

export function IconGift(props: IconProps) {
  return (
    <Icon {...props}>
      <rect x="4" y="9" width="16" height="4" rx="1" />
      <rect x="5" y="13" width="14" height="7" rx="1" />
      <path d="M12 9v11" />
      <path d="M12 9c-1-2.5-2.5-4-4.3-4a2 2 0 0 0 0 4M12 9c1-2.5 2.5-4 4.3-4a2 2 0 0 1 0 4" />
    </Icon>
  )
}

export function IconUtensils(props: IconProps) {
  return (
    <Icon {...props}>
      <path d="M5 3v6a2.5 2.5 0 0 0 5 0V3M7.5 9v12" />
      <path d="M16.5 3c-1.4 0-2.5 2-2.5 4.5s1.1 4.5 2.5 4.5v9" />
    </Icon>
  )
}

export function IconBabyFace(props: IconProps) {
  return (
    <Icon {...props}>
      <circle cx="12" cy="13" r="7" />
      <path d="M11 6.2c-.3-1.4.5-2.5 1.8-2.2" />
      <circle cx="9.2" cy="12" r="0.7" fill="currentColor" stroke="none" />
      <circle cx="14.8" cy="12" r="0.7" fill="currentColor" stroke="none" />
      <path d="M9.5 15.5c1 1 4 1 5 0" />
    </Icon>
  )
}

export function IconBanknote(props: IconProps) {
  return (
    <Icon {...props}>
      <rect x="2.5" y="6.5" width="19" height="11" rx="1.5" />
      <circle cx="12" cy="12" r="2.5" />
      <path d="M5.5 6.5v11M18.5 6.5v11" />
    </Icon>
  )
}

export function IconBriefcase(props: IconProps) {
  return (
    <Icon {...props}>
      <rect x="3" y="7.5" width="18" height="12" rx="2" />
      <path d="M8 7.5V6a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v1.5" />
      <path d="M3 12.5h18" />
    </Icon>
  )
}

export function IconTrendingUp(props: IconProps) {
  return (
    <Icon {...props}>
      <path d="M3 17l6-6 4 4 7-8" />
      <path d="M15 7h5v5" />
    </Icon>
  )
}

export function IconUndo(props: IconProps) {
  return (
    <Icon {...props}>
      <path d="M7 9 3 12.5 7 16" />
      <path d="M3 12.5h11.5a5.5 5.5 0 1 1 0 11H12" />
    </Icon>
  )
}

export function IconStar(props: IconProps) {
  return (
    <Icon {...props}>
      <path d="M12 3l2.6 5.9 6.4.6-4.8 4.3 1.4 6.2L12 16.9 6.4 20l1.4-6.2L3 9.5l6.4-.6L12 3Z" />
    </Icon>
  )
}
