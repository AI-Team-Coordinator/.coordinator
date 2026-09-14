interface LogoProps {
  size?: number | string
  className?: string
}

export function CoordinatorLogo({ size = 40, className }: LogoProps) {
  return (
    <img
      src="/logo.svg?v=3"
      width={size}
      height={size}
      className={className}
      alt=""
      draggable={false}
      aria-hidden
    />
  )
}
