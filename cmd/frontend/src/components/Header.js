"use client";

import Image from "next/image";

export default function Header() {
  return (
    <header className="w-full border-b border-border bg-card">
      <div className="mx-auto flex max-w-7xl items-center justify-between px-6 py-4">
        <div className="flex items-center gap-4">
            <Image
                src="/dtac_logo.webp"
                alt="Digitales Testfeld Air Cargo"
                width={120}
                height={40}
                className="h-10 w-auto"
                priority
            />
          <div className="hidden sm:block">
            <h1 className="text-lg font-semibold leading-tight text-foreground">
              eCommerce ONE Record Converter
            </h1>
            <p className="text-xs text-muted-foreground">
              CHI Deutschland Cargo Handling GmbH
            </p>
          </div>
        </div>

      </div>
    </header>
  );
}

