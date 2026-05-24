"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { removeToken, isAuthenticated, decodeToken } from "@/lib/auth";
import { Button } from "@/components/ui/button";
import { Avatar } from "@/components/ui/avatar";
// FIXED: Added the missing sub-components to the import statement
import { 
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem
} from "@/components/ui/dropdown-menu";

export function UserProfile() {
  const router = useRouter();
  const [user, setUser] = useState<{ id: number; email: string } | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const checkAuth = () => {
      if (isAuthenticated()) {
        const token = localStorage.getItem("token") || "";
        const decoded = decodeToken(token);
        if (decoded) {
          setUser(decoded);
        }
      } else {
        setUser(null);
      }
      setLoading(false);
    };

    checkAuth();
  }, []);

  const handleLogout = () => {
    removeToken();
    router.push("/login");
  };

  if (loading) {
    return <div>Loading...</div>;
  }

  if (!user) {
    return (
      <>
        <a href="/login" className="text-sm font-medium text-indigo-600 hover:text-indigo-500">
          Sign in
        </a>
      </>
    );
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="outline" size="sm" className="flex items-center space-x-2">
          <Avatar name={user.email} className="h-8 w-8" />
          <span className="hidden md:block">{user.email}</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent className="w-48" align="end" sideOffset={4}>
        <DropdownMenuItem onClick={handleLogout}>
          Sign out
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}