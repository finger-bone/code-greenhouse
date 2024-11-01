"use client";

import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { useAuthToken } from "@/providers/token-provider";
import axios from "axios";
import { Copy } from "lucide-react";
import { useEffect, useState } from "react";

export default function UserPage() {
  const token = useAuthToken();
  const [gitUserName, setGitUserName] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [isLoading, setIsLoading] = useState(true);
  const [isChangingPassword, setIsChangingPassword] = useState(false);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [dialogContent, setDialogContent] = useState({
    title: "",
    description: "",
  });

  useEffect(() => {
    axios
      .get(`/api/user/name?provider=${token.provider}`, {
        headers: {
          Authorization: `Bearer ${token.token}`,
        },
      })
      .then((response) => {
        setGitUserName(response.data.data.gitName);
        setIsLoading(false);
      })
      .catch((error) => {
        console.error("Failed to fetch user name:", error);
        setIsLoading(false);
        showDialog("Error", "Failed to fetch user name. Please try again.");
      });
  }, [token]);

  const copyUsername = () => {
    navigator.clipboard.writeText(gitUserName).then(() => {
      showDialog("Copied!", "Username copied to clipboard.");
    }).catch((err) => {
      console.error("Failed to copy text: ", err);
      showDialog("Error", "Failed to copy username. Please try again.");
    });
  };

  const handleChangePassword = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsChangingPassword(true);
    try {
      await axios.post(
        `/api/user/password?newPassword=${newPassword}`,
        {},
        {
          headers: {
            Authorization: `Bearer ${token.token}`,
          },
        },
      );
      showDialog(
        "Password changed",
        "Your password has been successfully updated.",
      );
      setNewPassword("");
    } catch (error) {
      console.error("Failed to change password:", error);
      showDialog("Error", "Failed to change password. Please try again.");
    } finally {
      setIsChangingPassword(false);
    }
  };

  const showDialog = (title: string, description: string) => {
    setDialogContent({ title, description });
    setDialogOpen(true);
  };

  return (
    <div className="container mx-auto m-16">
      <Card className="max-w-md mx-auto">
        <CardHeader>
          <CardTitle>User Profile</CardTitle>
          <CardDescription>Manage your account settings</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="username">Git Server Username</Label>
            <div className="flex items-center space-x-2">
              {isLoading ? <Skeleton className="h-9 w-full" /> : (
                <>
                  <Input
                    id="username"
                    value={gitUserName}
                    readOnly
                    className="flex-grow"
                  />
                  <Button
                    type="button"
                    variant="outline"
                    size="icon"
                    onClick={copyUsername}
                    aria-label="Copy username"
                  >
                    <Copy className="h-4 w-4" />
                  </Button>
                </>
              )}
            </div>
          </div>
          <div className="space-y-2">
            <Label htmlFor="provider">Provider</Label>
            <Input id="provider" value={token.provider ?? ""} readOnly />
          </div>
          <form onSubmit={handleChangePassword} className="space-y-2">
            <Label htmlFor="new-password">New Git Server Password</Label>
            <Input
              id="new-password"
              type="password"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              placeholder="Enter new password"
            />
            <Button
              type="submit"
              className="w-full"
              disabled={isChangingPassword}
            >
              {isChangingPassword ? "Changing Password..." : "Change Password"}
            </Button>
          </form>
        </CardContent>
      </Card>

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{dialogContent.title}</DialogTitle>
            <DialogDescription>{dialogContent.description}</DialogDescription>
          </DialogHeader>
        </DialogContent>
      </Dialog>
    </div>
  );
}
