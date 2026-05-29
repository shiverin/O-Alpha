export type AvatarProps = {
  name: string;
  size?: number;
  className?: string;
};

export const Avatar = ({ name, size = 40, className = "" }: AvatarProps) => {
  const initials = name
    .split(" ")
    .map((part) => part[0])
    .join("")
    .toUpperCase()
    .slice(0, 2);

  let hash = 0;
  for (let i = 0; i < name.length; i++) {
    hash = name.charCodeAt(i) + ((hash << 5) - hash);
  }
  const hue = Math.abs(hash) % 360;
  const bgColor = `hsl(${hue}, 70%, 40%)`;

  return (
    <div
      className={`flex h-[${size}px] w-[${size}px] items-center justify-center rounded-full bg-${bgColor} text-white text-sm font-medium ${className}`}
    >
      {initials}
    </div>
  );
};
