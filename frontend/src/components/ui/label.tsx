import type { HTMLAttributes } from "react";

export type LabelProps = HTMLAttributes<HTMLLabelElement> & {
  htmlFor?: string;
};

export const Label = ({ className = "", htmlFor, ...props }: LabelProps) => (
  <label
    className={[
      "text-sm font-medium text-gray-600 dark:text-gray-400",
      className,
    ].join(" ")}
    htmlFor={htmlFor}
    {...props}
  />
);

Label.displayName = "Label";
