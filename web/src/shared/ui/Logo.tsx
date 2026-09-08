import type { SVGProps } from 'react'

interface LogoProps extends SVGProps<SVGSVGElement> {
  size?: number | string
}

export function CoordinatorLogo({ size = 40, className, ...props }: LogoProps) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 48 48"
      width={size}
      height={size}
      fill="none"
      className={className}
      aria-hidden="true"
      {...props}
    >
      <rect width="48" height="48" rx="12" fill="#4F46E5" />
      <path
        d="M16 34 V24 C16 20 24 20 24 16 V12"
        stroke="white"
        strokeWidth="2.5"
        strokeLinecap="round"
      />
      <path
        d="M32 34 V24 C32 20 24 20 24 16"
        stroke="white"
        strokeWidth="2.5"
        strokeLinecap="round"
      />
      <circle cx="16" cy="34" r="3" fill="white" />
      <circle cx="32" cy="34" r="3" fill="white" />
      <circle cx="24" cy="12" r="3" fill="white" />
    </svg>
  )
}
