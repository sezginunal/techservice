import type { ReactNode } from "react";

export const wrenchPath =
  "M10.5 2.76V5.5h3V2.76A4.5 4.5 0 0 1 13.5 11.24V19.5A1.5 1.5 0 0 1 10.5 19.5V11.24A4.5 4.5 0 0 1 10.5 2.76z";

const dot = { fill: "currentColor", stroke: "none" };

const shapes = {
  home: (
    <>
      <path d="M3.75 10.25 12 3.75l8.25 6.5" />
      <path d="M5.75 8.75v10.5a1 1 0 0 0 1 1H10v-5.5h4v5.5h3.25a1 1 0 0 0 1-1V8.75" />
    </>
  ),
  services: (
    <g transform="rotate(45 12 12)">
      <path d={wrenchPath} />
    </g>
  ),
  customers: (
    <>
      <circle cx="9" cy="8" r="3.25" />
      <path d="M3.5 19.25c.5-3.1 2.75-5 5.5-5s5 1.9 5.5 5" />
      <path d="M15.5 4.9a3.25 3.25 0 0 1 0 6.2" />
      <path d="M17.25 14.5c1.8.5 2.95 2.1 3.25 4.75" />
    </>
  ),
  users: (
    <>
      <rect x="3.5" y="5" width="17" height="14" rx="2.5" />
      <circle cx="9" cy="10.75" r="2" />
      <path d="M6 16c.5-1.4 1.6-2.2 3-2.2s2.5.8 3 2.2" />
      <path d="M14.5 10h3M14.5 13.5h3" />
    </>
  ),
  spare: (
    <>
      <path d="M12 3.25 19.75 7.5v9L12 20.75 4.25 16.5v-9L12 3.25z" />
      <path d="M4.5 7.6 12 11.8l7.5-4.2" />
      <path d="M12 11.8v8.7" />
      <path d="m8.1 5.4 7.6 4.25" />
    </>
  ),
  audit: (
    <>
      <path d="M9 4.5H7a2 2 0 0 0-2 2v13a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2v-13a2 2 0 0 0-2-2h-2" />
      <rect x="9" y="3" width="6" height="3" rx="1" />
      <path d="M8.75 11h6.5M8.75 14.5h6.5M8.75 18h3.5" />
    </>
  ),
  settings: (
    <>
      <path d="M4 7h9M17 7h3M4 17h3M11 17h9" />
      <circle cx="15" cy="7" r="2" />
      <circle cx="9" cy="17" r="2" />
    </>
  ),
  menu: <path d="M4 7h16M4 12h16M4 17h16" />,
  close: <path d="M6 6l12 12M18 6 6 18" />,
  search: (
    <>
      <circle cx="11" cy="11" r="6.25" />
      <path d="m15.75 15.75 4 4" />
    </>
  ),
  plus: <path d="M12 5v14M5 12h14" />,
  minus: <path d="M5 12h14" />,
  chevronRight: <path d="m9.5 6 6 6-6 6" />,
  chevronLeft: <path d="m14.5 6-6 6 6 6" />,
  arrowRight: <path d="M5 12h14M13.5 6.5 19 12l-5.5 5.5" />,
  pencil: (
    <>
      <path d="M4 20h4L19 9a2.12 2.12 0 0 0-3-3L5 17z" />
      <path d="m14.5 7.5 3 3" />
    </>
  ),
  trash: (
    <>
      <path d="M4.5 7h15" />
      <path d="M9.5 7V4.75h5V7" />
      <path d="m6.5 7 .85 12.1a1.5 1.5 0 0 0 1.5 1.4h6.3a1.5 1.5 0 0 0 1.5-1.4L17.5 7" />
      <path d="M10.25 11v5.5M13.75 11v5.5" />
    </>
  ),
  imagePlus: (
    <>
      <path d="M20.5 12.5v5.75a2.25 2.25 0 0 1-2.25 2.25H5.75a2.25 2.25 0 0 1-2.25-2.25V5.75A2.25 2.25 0 0 1 5.75 3.5h7" />
      <circle cx="9" cy="9.5" r="1.75" />
      <path d="m20.5 16-4.25-4.25L7 20.5" />
      <path d="M18.5 3v6M15.5 6h6" />
    </>
  ),
  download: <path d="M12 4v11M7.5 10.75 12 15.25l4.5-4.5M5 20h14" />,
  calendar: (
    <>
      <rect x="3.75" y="5" width="16.5" height="15.25" rx="2.5" />
      <path d="M3.75 10h16.5M8.25 3v4M15.75 3v4" />
    </>
  ),
  check: <path d="m5 12.5 4.5 4.5 9.5-10" />,
  clock: (
    <>
      <circle cx="12" cy="12" r="8.5" />
      <path d="M12 7.5V12l3 2" />
    </>
  ),
  hourglass: (
    <>
      <path d="M6.5 3.5h11M6.5 20.5h11" />
      <path d="M8 3.5v2.25a4 4 0 0 0 1.7 3.27L12 10.75l2.3-1.73A4 4 0 0 0 16 5.75V3.5" />
      <path d="M8 20.5v-2.25a4 4 0 0 1 1.7-3.27L12 13.25l2.3 1.73a4 4 0 0 1 1.7 3.27v2.25" />
    </>
  ),
  inbox: (
    <>
      <path d="M3.5 13.5 6.2 5.6a1.5 1.5 0 0 1 1.4-1.1h8.8a1.5 1.5 0 0 1 1.4 1.1l2.7 7.9" />
      <path d="M3.5 13.5v4.75A1.75 1.75 0 0 0 5.25 20h13.5a1.75 1.75 0 0 0 1.75-1.75V13.5h-5l-1.25 2.25h-4.5L8.5 13.5z" />
    </>
  ),
  logout: (
    <>
      <path d="M14 4.5h3.75a1.75 1.75 0 0 1 1.75 1.75v11.5a1.75 1.75 0 0 1-1.75 1.75H14" />
      <path d="M9.5 8 5.5 12l4 4M5.5 12H15" />
    </>
  ),
  grip: (
    <>
      <circle cx="9" cy="6" r="1.35" {...dot} />
      <circle cx="15" cy="6" r="1.35" {...dot} />
      <circle cx="9" cy="12" r="1.35" {...dot} />
      <circle cx="15" cy="12" r="1.35" {...dot} />
      <circle cx="9" cy="18" r="1.35" {...dot} />
      <circle cx="15" cy="18" r="1.35" {...dot} />
    </>
  ),
  message: (
    <path d="M5.25 4.75h13.5a1.75 1.75 0 0 1 1.75 1.75v9a1.75 1.75 0 0 1-1.75 1.75H11l-4.5 3.25v-3.25H5.25A1.75 1.75 0 0 1 3.5 15.5v-9a1.75 1.75 0 0 1 1.75-1.75z" />
  ),
  history: (
    <>
      <path d="M4.25 12a7.75 7.75 0 1 0 2.3-5.5L4.25 8.75" />
      <path d="M4.25 4.5v4.25H8.5" />
      <path d="M12 8v4.25l2.75 1.75" />
    </>
  ),
  phone: (
    <>
      <rect x="6.75" y="2.75" width="10.5" height="18.5" rx="2.5" />
      <path d="M10.75 17.75h2.5" />
    </>
  ),
  mail: (
    <>
      <rect x="3.5" y="5.25" width="17" height="13.5" rx="2.25" />
      <path d="m4 7 8 5.75L20 7" />
    </>
  ),
  pin: (
    <>
      <path d="M12 20.75s-6.25-5.5-6.25-10.5a6.25 6.25 0 0 1 12.5 0c0 5-6.25 10.5-6.25 10.5z" />
      <circle cx="12" cy="10.25" r="2.25" />
    </>
  ),
  hash: <path d="M9.5 4 8 20M16 4l-1.5 16M4.5 9h15.5M4 15h15.5" />,
  device: (
    <>
      <rect x="3" y="4" width="18" height="12.5" rx="2.25" />
      <path d="M8.5 20h7M12 16.5V20" />
    </>
  ),
  layers: (
    <>
      <path d="M12 3.75 3.75 8.25 12 12.75l8.25-4.5L12 3.75z" />
      <path d="m3.75 12.25 8.25 4.5 8.25-4.5" />
      <path d="m3.75 16 8.25 4.5 8.25-4.5" />
    </>
  ),
  user: (
    <>
      <circle cx="12" cy="8" r="3.75" />
      <path d="M5 20c.6-3.6 3.4-5.75 7-5.75s6.4 2.15 7 5.75" />
    </>
  ),
  userPlus: (
    <>
      <circle cx="10" cy="8" r="3.5" />
      <path d="M3.75 19.5c.55-3.3 3-5.25 6.25-5.25 1.4 0 2.65.35 3.65 1" />
      <path d="M18 13.5v6M15 16.5h6" />
    </>
  ),
  chart: (
    <>
      <path d="M4 20h16" />
      <path d="M7 16.5v-5M12 16.5V6.5M17 16.5v-8" />
    </>
  ),
  list: (
    <>
      <path d="M9 6.5h11M9 12h11M9 17.5h11" />
      <circle cx="4.75" cy="6.5" r="1.1" {...dot} />
      <circle cx="4.75" cy="12" r="1.1" {...dot} />
      <circle cx="4.75" cy="17.5" r="1.1" {...dot} />
    </>
  ),
  eye: (
    <>
      <path d="M2.75 12S6.25 5.75 12 5.75 21.25 12 21.25 12 17.75 18.25 12 18.25 2.75 12 2.75 12z" />
      <circle cx="12" cy="12" r="2.75" />
    </>
  ),
  eyeOff: (
    <>
      <path d="M10.6 5.85A9.4 9.4 0 0 1 12 5.75c5.75 0 9.25 6.25 9.25 6.25a16.6 16.6 0 0 1-2.4 3.15M6.6 6.65C4.05 8.3 2.75 12 2.75 12S6.25 18.25 12 18.25a8.9 8.9 0 0 0 4.75-1.4" />
      <path d="M10.05 10.05a2.75 2.75 0 0 0 3.9 3.9" />
      <path d="M3.5 3.5l17 17" />
    </>
  ),
  lock: (
    <>
      <rect x="5" y="10.5" width="14" height="10" rx="2.25" />
      <path d="M8 10.5V7.75a4 4 0 0 1 8 0v2.75" />
    </>
  ),
  undo: (
    <>
      <path d="M9 14.5 4.5 10 9 5.5" />
      <path d="M4.5 10h10a5 5 0 0 1 0 10H11" />
    </>
  ),
  filter: <path d="M4 5.5h16l-6.25 7.25v5.5l-3.5 1.75v-7.25z" />,
  status: (
    <>
      <circle cx="12" cy="12" r="8.5" />
      <path d="m8.25 12.25 2.5 2.5 5-5" />
    </>
  ),
  printer: (
    <>
      <path d="M7 8.5V4.75h10V8.5" />
      <path d="M7 15.5H5.75A1.75 1.75 0 0 1 4 13.75v-3.5A1.75 1.75 0 0 1 5.75 8.5h12.5A1.75 1.75 0 0 1 20 10.25v3.5a1.75 1.75 0 0 1-1.75 1.75H17" />
      <rect x="7" y="13.5" width="10" height="6.75" rx="1.25" />
    </>
  ),
} satisfies Record<string, ReactNode>;

export type IconName = keyof typeof shapes;

export function Icon({ name, size = 16 }: { name: IconName; size?: number }) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth={1.75}
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
      focusable="false"
    >
      {shapes[name]}
    </svg>
  );
}
