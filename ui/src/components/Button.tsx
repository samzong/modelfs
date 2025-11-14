import React, { PropsWithChildren } from "react";

type Variant = "primary" | "secondary" | "outline" | "danger";
type Size = "sm" | "md";

export default function Button({ children, variant = "primary", size = "md", className = "", ...rest }: PropsWithChildren<{ variant?: Variant; size?: Size; className?: string } & React.ButtonHTMLAttributes<any>>) {
  const base = "inline-flex items-center justify-center rounded-lg font-medium transition-colors disabled:opacity-50 disabled:cursor-not-allowed";
  const sizes = size === "sm" ? "h-8 px-3 text-sm" : "h-9 px-4";
  const variants = variant === "primary" ? "bg-primary-600 text-white hover:bg-primary-700" : variant === "secondary" ? "bg-muted text-gray-800 hover:bg-gray-200" : variant === "outline" ? "border bg-white hover:bg-muted" : "bg-red-600 text-white hover:bg-red-700";
  return (
    <button className={`${base} ${sizes} ${variants} ${className}`} {...rest}>
      {children}
    </button>
  );
}
