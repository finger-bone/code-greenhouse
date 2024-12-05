import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { useAuthToken } from "@/providers/token-provider";
import { LogOut } from "lucide-react";
import { useState } from "react";
import { useNavigate } from "react-router-dom";
import LocaleToggle from "./locale-toggle";
import ModeToggle from "./mode-toggle";

export default function Header({
  items = [],
}: {
  items: {
    title: string;
    href: string;
  }[];
}) {
  const [isLogoutDialogOpen, setIsLogoutDialogOpen] = useState(false);
  const auth = useAuthToken();
  const navigate = useNavigate();

  const handleLogout = () => {
    auth.setProvider(null);
    auth.setToken(null);
    setIsLogoutDialogOpen(false);
    navigate("/login");
  };

  return (
    <header className="flex items-center justify-between p-4 bg-accent w-full">
      <div className="flex items-center gap-x-4">
        <nav className="flex items-center gap-x-16">
          {items.map((item) => (
            <a
              key={item.title}
              href={`/#${item.href}`}
              className="text-sm font-medium transition-colors hover:text-primary"
            >
              {item.title}
            </a>
          ))}
        </nav>
      </div>
      <div className="flex items-center gap-4">
        <LocaleToggle
          items={[
            { title: "English", identifier: "en" },
            { title: "简体中文", identifier: "zh_hans" },
          ]}
        />
        <ModeToggle />
        {auth.singleUser
          ? <></>
          : (
            <Dialog
              open={isLogoutDialogOpen}
              onOpenChange={setIsLogoutDialogOpen}
            >
              <DialogTrigger asChild>
                <Button
                  variant="outline"
                  size="icon"
                  aria-label="Logout"
                >
                  <LogOut className="h-4 w-4" />
                </Button>
              </DialogTrigger>
              <DialogContent>
                <DialogHeader>
                  <DialogTitle>Are you sure you want to log out?</DialogTitle>
                  <DialogDescription>
                    You will be redirected to the login page.
                  </DialogDescription>
                </DialogHeader>
                <DialogFooter>
                  <Button
                    variant="outline"
                    onClick={() => setIsLogoutDialogOpen(false)}
                  >
                    Cancel
                  </Button>
                  <Button onClick={handleLogout}>
                    Log out
                  </Button>
                </DialogFooter>
              </DialogContent>
            </Dialog>
          )}
      </div>
    </header>
  );
}
