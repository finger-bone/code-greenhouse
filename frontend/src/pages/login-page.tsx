import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { type ResponseWrapper, useFetch } from "@/hooks/useFetch";
import { useAuthState } from "@/providers/auth-state-provider";
import { useAuthToken } from "@/providers/token-provider";
import { SiGithub, SiGoogle } from "@icons-pack/react-simple-icons";
import { Loader2, Lock } from "lucide-react";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";

async function RequestLoginProvider(provider: string, state: string) {
  const redirect_uri = encodeURIComponent(`${window.location.origin}`);
  const query = new URLSearchParams({
    redirect_uri: redirect_uri,
    state: state,
    provider: provider,
  });
  const res = await fetch(`/api/auth/url?${query.toString()}`, {
    method: "GET",
  });
  return (await res.json()).data.url as string;
}

export default function LoginPage() {
  const { data: resp, loading, error } = useFetch<
    ResponseWrapper<{
      providers: Array<string>;
    }>
  >("/api/auth/providers");

  const [provider, setProvider] = useState<string | null>(null);
  const state = btoa(Date.now() + Math.random().toString(36).substring(2, 15));
  const auth = useAuthToken();
  const authState = useAuthState();
  const { t } = useTranslation();

  useEffect(() => {
    if (provider) {
      auth.setProvider(provider);
      authState.setValue(state);
      RequestLoginProvider(provider, state).then(url => {
        window.location.href = url;
      });
    }
  }, [provider]);

  const getProviderIcon = (providerName: string) => {
    switch (providerName.toLowerCase()) {
      case "github":
        return <SiGithub className="w-6 h-6" />;
      case "google":
        return <SiGoogle className="w-6 h-6" />;
      default:
        return <Lock className="w-6 h-6" />;
    }
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center h-screen">
        <Loader2 className="h-8 w-8 animate-spin" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex justify-center items-center h-screen">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle className="text-red-600">Error</CardTitle>
          </CardHeader>
          <CardContent>
            <p>{error.message}</p>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="flex justify-center items-center m-16">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>{t("login")}</CardTitle>
          <CardDescription>{t("login_prompt")}</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4">
            {resp?.data.providers.map((providerName) => (
              <Button
                key={providerName}
                onClick={() => setProvider(providerName)}
                className="w-full"
                variant="outline"
              >
                {getProviderIcon(providerName)}
                {t("login_by")} {providerName}
              </Button>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
