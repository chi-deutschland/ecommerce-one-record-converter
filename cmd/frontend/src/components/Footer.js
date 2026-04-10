export default function Footer() {
  const currentYear = new Date().getFullYear();

  return (
    <footer className="w-full border-t border-border bg-card">
      <div className="mx-auto flex max-w-7xl flex-col items-center justify-between gap-3 px-6 py-6 sm:flex-row">
        <div className="text-center text-xs text-muted-foreground sm:text-left">
          <p>
            &copy; {currentYear}{" "}
            <a
              href="https://chi-cargo.com/"
              target="_blank"
              rel="noopener noreferrer"
              className="underline hover:text-foreground transition-colors"
            >
              CHI Deutschland Cargo Handling GmbH
            </a>
            . All rights reserved.
          </p>
          <p className="mt-1">Licensed under the MIT License.</p>
        </div>
        <div className="text-center text-xs text-muted-foreground sm:text-right">
          <p>
            Part of the{" "}
            <a
              href="https://www.digitales-testfeld-air-cargo.de/"
              target="_blank"
              rel="noopener noreferrer"
              className="underline hover:text-foreground transition-colors"
            >
              Digitales Testfeld Air Cargo (DTAC)
            </a>{" "}
            research project.
          </p>
          <p className="mt-1">
            Powered by{" "}
            <a
              href="https://www.iata.org/one-record"
              target="_blank"
              rel="noopener noreferrer"
              className="underline hover:text-foreground transition-colors"
            >
              IATA ONE Record
            </a>
          </p>
        </div>
      </div>
    </footer>
  );
}

