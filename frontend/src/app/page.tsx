"use client"; // Mark this file as a Client Component

import React, { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Component } from "../components/component";
import { Login } from "../components/login";

export default function Home() {
  const [currentUserName, setCurrentUserName] = useState<string | null>(null);
  const router = useRouter();

  useEffect(() => {
    const storedUserName = localStorage.getItem("currentUserName");
    if (storedUserName) {
      setCurrentUserName(storedUserName);
    }
  }, [router]);

  return (
    <div className="grid grid-rows-[20px_1fr_20px] items-center justify-items-center min-h-screen p-8 pb-20 gap-16 sm:p-20 font-[family-name:var(--font-geist-sans)]">
      <main className="flex flex-col gap-8 row-start-2 items-center sm:items-start">
        {currentUserName ? (
          <Component currentUserName={currentUserName} />
        ) : (
          <Login onLoginSuccess={setCurrentUserName} />
        )}
      </main>
    </div>
  );
}
