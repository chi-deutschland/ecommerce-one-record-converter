"use client";

import { useContext, useState } from "react";
import { useDropzone } from "react-dropzone";
import { Input } from "@/components/ui/input";
import { GlobalContext } from "@/context/GlobalContext";
import { useRouter } from "next/navigation";

const STEPS = [
  { label: "Upload", number: 1 },
  { label: "Processing", number: 2 },
];

function StepIndicator({ currentStep }) {
  return (
    <nav aria-label="Progress" className="flex items-center justify-center gap-2">
      {STEPS.map((step, idx) => {
        const isActive = step.number <= currentStep;
        return (
          <div key={step.number} className="flex items-center gap-2">
            <div className="flex flex-col items-center">
              <div
                className={`flex h-10 w-10 items-center justify-center rounded-full text-sm font-semibold transition-colors ${
                  isActive
                    ? "bg-primary text-primary-foreground"
                    : "bg-muted text-muted-foreground"
                }`}
              >
                {step.number}
              </div>
              <span className="mt-1 text-xs text-muted-foreground">{step.label}</span>
            </div>
            {idx < STEPS.length - 1 && (
              <div
                className={`mb-5 h-0.5 w-20 sm:w-32 ${
                  currentStep > step.number ? "bg-primary" : "bg-muted"
                }`}
              />
            )}
          </div>
        );
      })}
    </nav>
  );
}

export default function UploadPage() {
  const { neoneServerBaseAddress, setNeoneServerBaseAddress } = useContext(GlobalContext);
  const { neoneAuthToken, setNeoneAuthToken } = useContext(GlobalContext);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);
  const [file, setFile] = useState(null);
  const router = useRouter();

  const onDrop = (acceptedFiles) => {
    setError(null);
    if (acceptedFiles.length > 0) {
      setFile(acceptedFiles[0]);
    }
  };

  const handleUpload = async () => {
    setError(null);

    if (!neoneServerBaseAddress || !neoneServerBaseAddress.trim()) {
      setError("Please enter the NE:ONE Server base address.");
      return;
    }
    if (!neoneAuthToken || !neoneAuthToken.trim()) {
      setError("Please enter the NE:ONE Server authentication token.");
      return;
    }
    if (!file) {
      setError("Please select an Excel file (.xlsx) to upload.");
      return;
    }

    setIsLoading(true);
    const formData = new FormData();
    formData.append("neoneServerBaseAddress", neoneServerBaseAddress);
    formData.append("file", file);

    try {
      const response = await fetch("/upload", {
        method: "POST",
        credentials: "include",
        headers: {
          Authorization: `Bearer ${neoneAuthToken}`,
        },
        body: formData,
      });

      if (!response.ok) {
        const responseText = await response.text();
        setError(
          `Upload failed (HTTP ${response.status}). ${responseText || "Please check your configuration and try again."}`
        );
        setIsLoading(false);
        return;
      }

      setTimeout(() => {
        router.push("/done");
      }, 1000);
    } catch (err) {
      setError("A network error occurred. Please check your connection and try again.");
      setIsLoading(false);
    }
  };

  const { getRootProps, getInputProps, isDragActive, acceptedFiles } = useDropzone({
    maxFiles: 1,
    accept: {
      "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": [".xlsx"],
    },
    onDrop,
  });

  return (
    <div className="flex flex-col items-center px-4 py-10">
      <div className="mb-10">
        <StepIndicator currentStep={1} />
      </div>

      <div className="w-full max-w-3xl">
        <div className="rounded-xl border border-border bg-card p-6 shadow-sm sm:p-8">
          <h2 className="mb-1 text-xl font-semibold text-foreground">
            Upload Shipment Data
          </h2>
          <p className="mb-6 text-sm text-muted-foreground">
            Configure your NE:ONE Server connection and upload an Excel file containing
            eCommerce shipment data to convert it into IATA ONE Record format.
          </p>

          {error && (
            <div
              className="mb-6 rounded-lg border border-red-300 bg-red-50 px-4 py-3 text-sm text-red-800 dark:border-red-800 dark:bg-red-950 dark:text-red-200"
              role="alert"
            >
              {error}
            </div>
          )}

          <div className="space-y-6">
            {/* Server Configuration */}
            <fieldset className="space-y-4">
              <legend className="text-sm font-medium text-foreground">
                Server Configuration
              </legend>
              <div>
                <label
                  htmlFor="neoneServerBaseAddress"
                  className="mb-1 block text-sm text-muted-foreground"
                >
                  NE:ONE Server Base Address
                </label>
                <Input
                  id="neoneServerBaseAddress"
                  name="neoneServerBaseAddress"
                  type="url"
                  placeholder="https://your-neone-server.example.com"
                  value={neoneServerBaseAddress || ""}
                  onChange={(e) => setNeoneServerBaseAddress(e.target.value)}
                />
              </div>
              <div>
                <label
                  htmlFor="neoneAuthToken"
                  className="mb-1 block text-sm text-muted-foreground"
                >
                  Authentication Token
                </label>
                <Input
                  id="neoneAuthToken"
                  name="neoneAuthToken"
                  type="password"
                  placeholder="Bearer token for NE:ONE Server"
                  value={neoneAuthToken || ""}
                  onChange={(e) => setNeoneAuthToken(e.target.value)}
                />
              </div>
            </fieldset>

            {/* File Upload */}
            <fieldset>
              <legend className="mb-2 text-sm font-medium text-foreground">
                Shipment Data File
              </legend>
              <div
                {...getRootProps()}
                className={`flex cursor-pointer flex-col items-center justify-center rounded-lg border-2 border-dashed px-6 py-16 text-center transition-colors ${
                  isDragActive
                    ? "border-primary bg-primary/5"
                    : "border-muted-foreground/25 hover:border-primary/50 hover:bg-muted/50"
                }`}
              >
                <input {...getInputProps()} />
                <svg
                  className="mb-3 h-10 w-10 text-muted-foreground"
                  xmlns="http://www.w3.org/2000/svg"
                  fill="none"
                  viewBox="0 0 24 24"
                  strokeWidth={1.5}
                  stroke="currentColor"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    d="M3 16.5v2.25A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75V16.5m-13.5-9L12 3m0 0l4.5 4.5M12 3v13.5"
                  />
                </svg>
                {isDragActive ? (
                  <p className="text-sm text-primary">Drop your file here&hellip;</p>
                ) : (
                  <>
                    <p className="text-sm text-muted-foreground">
                      Drag &amp; drop your Excel file here, or{" "}
                      <span className="font-medium text-primary underline">browse</span>
                    </p>
                    <p className="mt-1 text-xs text-muted-foreground/70">
                      Accepted format: .xlsx
                    </p>
                  </>
                )}
              </div>
              {acceptedFiles.length > 0 && (
                <div className="mt-3 rounded-md border border-border bg-muted/50 px-4 py-2 text-sm text-foreground">
                  <span className="font-medium">Selected:</span>{" "}
                  {acceptedFiles[0].path} ({(acceptedFiles[0].size / 1024).toFixed(1)} KB)
                </div>
              )}
            </fieldset>
          </div>

          <div className="mt-8 flex justify-end">
            <button
              onClick={handleUpload}
              disabled={isLoading}
              className="inline-flex items-center gap-2 rounded-lg bg-primary px-6 py-2.5 text-sm font-medium text-primary-foreground shadow-sm transition-colors hover:bg-primary/90 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
            >
              {isLoading ? (
                <>
                  <svg
                    className="h-4 w-4 animate-spin"
                    xmlns="http://www.w3.org/2000/svg"
                    fill="none"
                    viewBox="0 0 24 24"
                  >
                    <circle
                      className="opacity-25"
                      cx="12"
                      cy="12"
                      r="10"
                      stroke="currentColor"
                      strokeWidth="4"
                    />
                    <path
                      className="opacity-75"
                      fill="currentColor"
                      d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
                    />
                  </svg>
                  Uploading&hellip;
                </>
              ) : (
                "Upload & Convert"
              )}
            </button>
          </div>
        </div>
      </div>

      {isLoading && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-background/80 backdrop-blur-sm">
          <div className="flex flex-col items-center gap-4 rounded-xl bg-card p-8 shadow-lg border border-border">
            <svg
              className="h-10 w-10 animate-spin text-primary"
              xmlns="http://www.w3.org/2000/svg"
              fill="none"
              viewBox="0 0 24 24"
            >
              <circle
                className="opacity-25"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                strokeWidth="4"
              />
              <path
                className="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
              />
            </svg>
            <p className="text-sm font-medium text-foreground">
              Processing your shipment data&hellip;
            </p>
          </div>
        </div>
      )}
    </div>
  );
}
