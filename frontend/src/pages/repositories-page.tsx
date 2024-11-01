import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import useGql from "@/hooks/useGql";
import { useAuthToken } from "@/providers/token-provider";
import { File } from "lucide-react";

interface Repository {
  repositoryId: string;
  subject: string;
  provider: string;
  challengeFolderName: string;
  startpoint: string;
  stage: number;
  totalStages: number;
  createTime: string;
  updateTime: string;
}

export default function RepositoriesPage() {
  const token = useAuthToken();
  const { data, loading, error } = useGql<{
    repositories: Repository[];
  }>(`
    query Repositories {
      repositories(subject: "${token.subject}", provider: "${token.provider}") {
        repositoryId
        subject
        provider
        challengeFolderName
        startpoint
        stage
        totalStages
        createTime
        updateTime
      }
    }
  `);

  if (loading) {
    return (
      <div className="container mx-auto p-4">
        <h1 className="text-2xl font-bold mb-4">Repositories</h1>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {[...Array(6)].map((_, i) => (
            <Card key={i} className="w-full">
              <CardHeader>
                <Skeleton className="h-6 w-3/4 mb-2" />
                <Skeleton className="h-4 w-1/2" />
              </CardHeader>
              <CardContent>
                <Skeleton className="h-4 w-full mb-2" />
                <Skeleton className="h-4 w-3/4" />
              </CardContent>
              <CardFooter>
                <Skeleton className="h-8 w-24" />
              </CardFooter>
            </Card>
          ))}
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="container mx-auto p-4">
        <Alert variant="destructive">
          <AlertTitle>Error</AlertTitle>
          <AlertDescription>
            An error occurred while fetching repositories. Please try again
            later.
          </AlertDescription>
        </Alert>
      </div>
    );
  }

  return (
    <div className="container mx-auto py-8">
      <h1 className="text-3xl font-bold mb-8 flex items-center gap-2">
        Repositories
      </h1>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {data?.repositories.map((repo) => (
          <a
            href={`/#/repository/${repo.repositoryId}`}
            key={repo.repositoryId}
          >
            <Card className="w-full h-full hover:shadow-lg transition-shadow duration-300">
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <File className="h-5 w-5" />
                  {repo.challengeFolderName}
                </CardTitle>
                <CardDescription>{repo.repositoryId}</CardDescription>
              </CardHeader>
              <CardContent>
                <p className="text-sm text-muted-foreground mb-2">
                  Startpoint: {repo.startpoint}
                </p>
                <p className="text-sm text-muted-foreground">
                  Created: {new Date(repo.createTime).toLocaleDateString()}
                </p>
              </CardContent>
              <CardFooter>
                <Badge variant="secondary">
                  Stage {repo.stage} / {repo.totalStages}
                </Badge>
              </CardFooter>
            </Card>
          </a>
        ))}
      </div>
    </div>
  );
}
