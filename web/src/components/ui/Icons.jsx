const baseProps = {
  fill: "none",
  stroke: "currentColor",
  strokeLinecap: "round",
  strokeLinejoin: "round",
  strokeWidth: "1.8",
  viewBox: "0 0 24 24",
};

function Icon({ children, className = "h-4 w-4" }) {
  return (
    <svg aria-hidden="true" className={className} {...baseProps}>
      {children}
    </svg>
  );
}

export function MenuIcon({ className }) {
  return (
    <Icon className={className}>
      <path d="M4 7h16" />
      <path d="M4 12h16" />
      <path d="M4 17h16" />
    </Icon>
  );
}

export function CloseIcon({ className }) {
  return (
    <Icon className={className}>
      <path d="m6 6 12 12" />
      <path d="m18 6-12 12" />
    </Icon>
  );
}

export function OverviewIcon({ className }) {
  return (
    <Icon className={className}>
      <rect x="3.5" y="3.5" width="7" height="7" rx="1.5" />
      <rect x="13.5" y="3.5" width="7" height="11" rx="1.5" />
      <rect x="3.5" y="13.5" width="7" height="7" rx="1.5" />
      <rect x="13.5" y="17.5" width="7" height="3" rx="1.5" />
    </Icon>
  );
}

export function ServicesIcon({ className }) {
  return (
    <Icon className={className}>
      <rect x="4" y="5" width="16" height="5" rx="1.5" />
      <rect x="4" y="14" width="16" height="5" rx="1.5" />
      <path d="M8 7.5h.01" />
      <path d="M8 16.5h.01" />
    </Icon>
  );
}

export function DiscoveryIcon({ className }) {
  return (
    <Icon className={className}>
      <circle cx="12" cy="12" r="2.5" />
      <path d="M4 12a8 8 0 0 1 8-8" />
      <path d="M20 12a8 8 0 0 0-8-8" />
      <path d="M4 12a8 8 0 0 0 8 8" />
      <path d="M20 12a8 8 0 0 1-8 8" />
    </Icon>
  );
}

export function DevicesIcon({ className }) {
  return (
    <Icon className={className}>
      <rect x="4" y="5" width="16" height="10" rx="2" />
      <path d="M8 19h8" />
      <path d="M10 15v4" />
      <path d="M14 15v4" />
    </Icon>
  );
}

export function BookmarkIcon({ className }) {
  return (
    <Icon className={className}>
      <path d="M7 4.5h10a1.5 1.5 0 0 1 1.5 1.5v13l-6.5-3-6.5 3V6A1.5 1.5 0 0 1 7 4.5Z" />
    </Icon>
  );
}

export function ActivityIcon({ className }) {
  return (
    <Icon className={className}>
      <path d="M4 12h4l2.5-5 3 10 2.5-5H20" />
    </Icon>
  );
}

export function PlusIcon({ className }) {
  return (
    <Icon className={className}>
      <path d="M12 5v14" />
      <path d="M5 12h14" />
    </Icon>
  );
}

export function RefreshIcon({ className }) {
  return (
    <Icon className={className}>
      <path d="M20 11a8 8 0 0 0-14.9-3" />
      <path d="M4 4v4h4" />
      <path d="M4 13a8 8 0 0 0 14.9 3" />
      <path d="M20 20v-4h-4" />
    </Icon>
  );
}

export function ArrowUpRightIcon({ className }) {
  return (
    <Icon className={className}>
      <path d="M7 17 17 7" />
      <path d="M8 7h9v9" />
    </Icon>
  );
}

export function ChevronDownIcon({ className }) {
  return (
    <Icon className={className}>
      <path d="m6 9 6 6 6-6" />
    </Icon>
  );
}

export function ShieldIcon({ className }) {
  return (
    <Icon className={className}>
      <path d="M12 3.5 5 6v5.5c0 4.2 2.7 8 7 9.5 4.3-1.5 7-5.3 7-9.5V6l-7-2.5Z" />
      <path d="m9.5 12 1.8 1.8 3.7-3.8" />
    </Icon>
  );
}

export function DatabaseIcon({ className }) {
  return (
    <Icon className={className}>
      <ellipse cx="12" cy="6" rx="7" ry="3" />
      <path d="M5 6v6c0 1.7 3.1 3 7 3s7-1.3 7-3V6" />
      <path d="M5 12v6c0 1.7 3.1 3 7 3s7-1.3 7-3v-6" />
    </Icon>
  );
}

export function SparklesIcon({ className }) {
  return (
    <Icon className={className}>
      <path d="m12 3 1.5 4.5L18 9l-4.5 1.5L12 15l-1.5-4.5L6 9l4.5-1.5L12 3Z" />
      <path d="m19 14 .8 2.2L22 17l-2.2.8L19 20l-.8-2.2L16 17l2.2-.8L19 14Z" />
      <path d="m5 14 .8 2.2L8 17l-2.2.8L5 20l-.8-2.2L2 17l2.2-.8L5 14Z" />
    </Icon>
  );
}

export function ClockIcon({ className }) {
  return (
    <Icon className={className}>
      <circle cx="12" cy="12" r="8.5" />
      <path d="M12 7.5v5l3.5 2" />
    </Icon>
  );
}

export function TokenIcon({ className }) {
  return (
    <Icon className={className}>
      <path d="M14.5 7.5a3.5 3.5 0 1 1 0 7h-5" />
      <path d="M9.5 16.5a3.5 3.5 0 1 1 0-7h5" />
      <path d="M9 12h6" />
    </Icon>
  );
}

export function FolderIcon({ className }) {
  return (
    <Icon className={className}>
      <path d="M4 8.5A1.5 1.5 0 0 1 5.5 7H10l1.5 2H18.5A1.5 1.5 0 0 1 20 10.5v7A1.5 1.5 0 0 1 18.5 19h-13A1.5 1.5 0 0 1 4 17.5v-9Z" />
    </Icon>
  );
}

export function SearchIcon({ className }) {
  return (
    <Icon className={className}>
      <circle cx="11" cy="11" r="6.5" />
      <path d="m16 16 4 4" />
    </Icon>
  );
}

export function PinIcon({ className }) {
  return (
    <Icon className={className}>
      <path d="M12 4v6" />
      <path d="m8 8 4-4 4 4" />
      <path d="M9 14h6" />
      <path d="m12 14-2 6" />
    </Icon>
  );
}

export function EditIcon({ className }) {
  return (
    <Icon className={className}>
      <path d="M4 20h4l10-10-4-4L4 16v4Z" />
      <path d="m12.5 7.5 4 4" />
    </Icon>
  );
}

export function TrashIcon({ className }) {
  return (
    <Icon className={className}>
      <path d="M5 7h14" />
      <path d="M9 7V5.5A1.5 1.5 0 0 1 10.5 4h3A1.5 1.5 0 0 1 15 5.5V7" />
      <path d="M7 7v11a2 2 0 0 0 2 2h6a2 2 0 0 0 2-2V7" />
      <path d="M10 11v5" />
      <path d="M14 11v5" />
    </Icon>
  );
}

export function DownloadIcon({ className }) {
  return (
    <Icon className={className}>
      <path d="M12 4v10" />
      <path d="m8 10 4 4 4-4" />
      <path d="M5 19h14" />
    </Icon>
  );
}

export function UploadIcon({ className }) {
  return (
    <Icon className={className}>
      <path d="M12 20V10" />
      <path d="m8 14 4-4 4 4" />
      <path d="M5 5h14" />
    </Icon>
  );
}

export function ArrowUpIcon({ className }) {
  return (
    <Icon className={className}>
      <path d="m7 13 5-5 5 5" />
    </Icon>
  );
}

export function ArrowDownIcon({ className }) {
  return (
    <Icon className={className}>
      <path d="m7 11 5 5 5-5" />
    </Icon>
  );
}
