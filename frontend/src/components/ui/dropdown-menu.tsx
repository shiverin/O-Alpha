import * as RadixUI from "@radix-ui/react-dropdown-menu";
//import { Slot } from "@radix-ui/react-slot";

import { cn } from "@/lib/utils";

const DropdownMenu = RadixUI.Root;
const DropdownMenuTrigger = RadixUI.Trigger;
const DropdownMenuContent = RadixUI.Content;
const DropdownMenuItem = RadixUI.Item;
const DropdownMenuCheckboxItem = RadixUI.CheckboxItem;
const DropdownMenuRadioGroup = RadixUI.RadioGroup;
const DropdownMenuRadioItem = RadixUI.RadioItem;
const DropdownMenuLabel = RadixUI.Label;
const DropdownMenuSeparator = RadixUI.Separator;
const DropdownMenuShortcut = ({
  className,
  ...props
}: React.HTMLAttributes<HTMLSpanElement>) => {
  return (
    <span
      className={cn("ml-auto text-xs tracking-widest opacity-60", className)}
      {...props}
    />
  );
};
DropdownMenuShortcut.displayName = "DropdownMenuShortcut";
const DropdownMenuGroup = RadixUI.Group;
const DropdownMenuSub = RadixUI.Sub;
const DropdownMenuSubTrigger = RadixUI.SubTrigger;
const DropdownMenuSubContent = RadixUI.SubContent;

export {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuCheckboxItem,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuShortcut,
  DropdownMenuGroup,
  DropdownMenuSub,
  DropdownMenuSubTrigger,
  DropdownMenuSubContent,
};
