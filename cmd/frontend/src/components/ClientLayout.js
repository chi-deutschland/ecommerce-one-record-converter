"use client";

import { ThemeProvider } from "@/components/theme-provider";
import { GlobalProvider } from "@/context/GlobalContext";
import Header from "@/components/Header";
import Footer from "@/components/Footer";

export default function ClientLayout({ children }) {
  return (
    <ThemeProvider
      attribute="class"
      defaultTheme="system"
      enableSystem
      disableTransitionOnChange
    >
      <GlobalProvider>
        <div className="flex min-h-screen flex-col">
          <Header />
          <main className="flex-1">{children}</main>
          <Footer />
        </div>
      </GlobalProvider>
    </ThemeProvider>
  );
}

