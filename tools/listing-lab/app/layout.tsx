import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Listing Lab · MLIP Dev",
  description: "Dev-only seller simulator for the MLIP listing pipeline",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
