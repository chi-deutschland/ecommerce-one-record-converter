"use client";

const STEPS = [
  { label: "Upload", number: 1 },
  { label: "Processing", number: 2 },
];

function StepIndicator({ currentStep }) {
  return (
    <nav aria-label="Progress" className="flex items-center justify-center gap-2">
      {STEPS.map((step, idx) => {
        const isComplete = step.number <= currentStep;
        return (
          <div key={step.number} className="flex items-center gap-2">
            <div className="flex flex-col items-center">
              <div
                className={`flex h-10 w-10 items-center justify-center rounded-full text-sm font-semibold transition-colors ${
                  isComplete
                    ? "bg-green-600 text-white"
                    : "bg-muted text-muted-foreground"
                }`}
              >
                ✓
              </div>
              <span className="mt-1 text-xs text-muted-foreground">{step.label}</span>
            </div>
            {idx < STEPS.length - 1 && (
              <div className="mb-5 h-0.5 w-20 bg-green-600 sm:w-32" />
            )}
          </div>
        );
      })}
    </nav>
  );
}

export default function DonePage() {
  return (
    <div className="flex flex-col items-center px-4 py-10">
      <div className="mb-10">
        <StepIndicator currentStep={2} />
      </div>

      <div className="w-full max-w-2xl">
        <div className="rounded-xl border border-border bg-card p-8 shadow-sm">
          <div className="flex items-start gap-4">
            <div className="flex h-14 w-14 flex-shrink-0 items-center justify-center rounded-full border-2 border-green-500 bg-green-500/10">
              <svg
                xmlns="http://www.w3.org/2000/svg"
                className="h-8 w-8 text-green-500"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                strokeWidth={2}
              >
                <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
              </svg>
            </div>
            <div>
              <h2 className="text-2xl font-semibold text-foreground">
                Data Processing Started
              </h2>
              <p className="mt-1 text-sm text-muted-foreground">
                Your shipment data has been accepted and is being converted.
              </p>
            </div>
          </div>

          <hr className="my-6 border-border" />

          <div className="space-y-4 text-sm text-muted-foreground">
            <div className="flex items-start gap-3">
              <div className="mt-0.5 flex h-5 w-5 flex-shrink-0 items-center justify-center rounded-full bg-primary/10 text-primary">
                <span className="text-xs font-bold">1</span>
              </div>
              <p>
                Your eCommerce data is being converted into IATA ONE Record Logistics
                Objects and posted to the NE:ONE Server.
              </p>
            </div>
            <div className="flex items-start gap-3">
              <div className="mt-0.5 flex h-5 w-5 flex-shrink-0 items-center justify-center rounded-full bg-primary/10 text-primary">
                <span className="text-xs font-bold">2</span>
              </div>
              <p>
                Notifications will be sent out for each Box-level Piece created on the
                server.
              </p>
            </div>
            <div className="flex items-start gap-3">
              <div className="mt-0.5 flex h-5 w-5 flex-shrink-0 items-center justify-center rounded-full bg-primary/10 text-primary">
                <span className="text-xs font-bold">3</span>
              </div>
              <p>
                Once processing is complete, the data will be available in your NE:ONE
                Server instance.
              </p>
            </div>
          </div>

          <div className="mt-8 flex justify-center">
            <a
              href="/upload"
              className="inline-flex items-center gap-2 rounded-lg border border-border bg-muted px-5 py-2.5 text-sm font-medium text-foreground shadow-sm transition-colors hover:bg-accent"
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                className="h-4 w-4"
                fill="none"
                viewBox="0 0 24 24"
                strokeWidth={1.5}
                stroke="currentColor"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M9 15L3 9m0 0l6-6M3 9h12a6 6 0 010 12h-3"
                />
              </svg>
              Upload Another File
            </a>
          </div>
        </div>
      </div>
    </div>
  );
}