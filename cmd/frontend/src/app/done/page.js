"use client";
import React from 'react';

export default function IntegrationPage() {
  return (
    <div>
      <div className="flex items-center justify-center mt-24">
        <div className="flex items-center">
          <div className="flex items-center justify-center w-10 h-10 rounded-full bg-green-600 text-xl">✓</div>
          <div className="w-32 h-1 bg-green-600 mx-2"></div>
          <div className="flex items-center justify-center w-10 h-10 rounded-full bg-green-600 text-xl">✓</div>
        </div>
      </div>
      <div className="flex justify-center items-center min-h-screen position-relative">
        <div className="flex flex-col items-center gap-5">
          <div className="container p-5 inline-block rounded-lg">
            <div className="flex items-center mb-16 mt-8">
              <div
                className="flex justify-center items-center rounded-full border-4 border-[#36b47E] h-[75px] w-[75px]">
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  className="h-16 w-16"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                  strokeWidth={2}
                  style={{color: '#36b47E'}}
                >
                  <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7"/>
                </svg>
              </div>
              <h1 className="ml-4 text-3xl">Your data is being processed</h1>
            </div>
            <div className="flex items-top">
              <h2 className="mt-5 ml-4">Soon it will be available in the NE:ONE Server.</h2>
            </div>
            <div>
              <h2 className="mt-5 ml-4">Notifications will be sent out accordingly for each Box-level Piece
                created.</h2>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}