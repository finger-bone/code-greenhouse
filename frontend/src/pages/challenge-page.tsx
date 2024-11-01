"use client";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
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
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { Skeleton } from "@/components/ui/skeleton";
import { ResponseWrapper } from "@/hooks/useFetch";
import useGql from "@/hooks/useGql";
import axios from "axios";
import { AlertCircle, Book, User } from "lucide-react";
import { useState } from "react";
import { useNavigate } from "react-router-dom";

interface Challenge {
  folderName: string;
  basic: {
    author: string;
    source: string;
    title: string;
    description: string[];
  };
}

interface ChallengeStartPoints {
  challenge: {
    startPoints: {
      name: string;
      description: string[];
    }[];
  };
}

const CHALLENGES_QUERY = `
query Challenges {
  challenges {
    folderName
    basic {
      author
      source
      title
      description
    }
  }
}
`;

function ChallengeCard({ challenge }: { challenge: Challenge }) {
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [selectedStartPoint, setSelectedStartPoint] = useState("");
  const [dialogState, setDialogState] = useState<
    "select" | "creating" | "created"
  >("select");
  const [repoId, setRepoId] = useState("");

  const { data, loading, error } = useGql<ChallengeStartPoints>(`
      query Challenge {
          challenge(folderName: "${challenge.folderName}") {
              startPoints {
                  name
                  description
              }
          }
      }
      `);

  const handleStartChallenge = async () => {
    setDialogState("creating");
    try {
      // Simulating API call to create repository
      const params = new URLSearchParams();
      params.append("folder", challenge.folderName);
      params.append("startpoint", selectedStartPoint);
      const resp = (await axios.post(`/api/repo/project?${params.toString()}`))
        .data as ResponseWrapper<{
          repositoryId: string;
        }>;
      const newRepoId = resp.data.repositoryId;
      setRepoId(newRepoId);
      setDialogState("created");
    } catch (error) {
      console.error("Failed to create repository:", error);
      setDialogState("select");
    }
  };

  const navigate = useNavigate();
  const handleGoToRepository = () => {
    navigate(`/repository/${repoId}`);
  };

  return (
    <Card className="h-full flex flex-col">
      <CardHeader>
        <CardTitle className="flex items-center justify-between">
          <span>{challenge.basic.title}</span>
          <Badge variant="secondary">{challenge.basic.source}</Badge>
        </CardTitle>
        <CardDescription className="flex items-center gap-2">
          <User className="h-4 w-4" />
          {challenge.basic.author}
        </CardDescription>
      </CardHeader>
      <CardContent className="flex-grow">
        <ul className="list-disc pl-5 space-y-1 mb-4">
          {challenge.basic.description.map((desc, index) => (
            <li key={index} className="text-sm text-muted-foreground">
              {desc}
            </li>
          ))}
        </ul>
      </CardContent>
      <CardContent className="pt-0">
        <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
          <DialogTrigger asChild>
            <Button className="w-full">Start Challenge</Button>
          </DialogTrigger>
          <DialogContent className="max-w-3xl">
            {dialogState === "select" && (
              <>
                <DialogHeader>
                  <DialogTitle>Choose a Start Point</DialogTitle>
                  <DialogDescription>
                    Select a starting point for your challenge.
                  </DialogDescription>
                </DialogHeader>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mt-4">
                  {data?.challenge?.startPoints.map((startPoint, index) => (
                    <Card
                      key={index}
                      className={`cursor-pointer transition-all ${
                        selectedStartPoint === startPoint.name
                          ? "ring-2 ring-primary"
                          : ""
                      }`}
                      onClick={() => setSelectedStartPoint(startPoint.name)}
                    >
                      <CardHeader>
                        <CardTitle>{startPoint.name}</CardTitle>
                      </CardHeader>
                      <CardContent>
                        <ul className="list-disc pl-5 space-y-1">
                          {startPoint.description.map((desc, idx) => (
                            <li
                              key={idx}
                              className="text-sm text-muted-foreground"
                            >
                              {desc}
                            </li>
                          ))}
                        </ul>
                      </CardContent>
                    </Card>
                  ))}
                </div>
                <DialogFooter className="mt-4">
                  <Button
                    onClick={handleStartChallenge}
                    disabled={!selectedStartPoint}
                  >
                    Start Challenge
                  </Button>
                </DialogFooter>
              </>
            )}
            {dialogState === "creating" && (
              <DialogDescription>Creating repository...</DialogDescription>
            )}
            {dialogState === "created" && (
              <>
                <DialogHeader>
                  <DialogTitle>Repository Created</DialogTitle>
                  <DialogDescription>
                    Your repository has been created with ID: {repoId}
                  </DialogDescription>
                </DialogHeader>
                <DialogFooter>
                  <Button onClick={() => setIsDialogOpen(false)}>Done</Button>
                  <Button onClick={handleGoToRepository}>
                    Go to Repository
                  </Button>
                </DialogFooter>
              </>
            )}
          </DialogContent>
        </Dialog>
      </CardContent>
    </Card>
  );
}

export default function ChallengePage() {
  const { data, loading, error } = useGql<{
    challenges: Challenge[];
  }>(CHALLENGES_QUERY);

  if (error) {
    return (
      <Alert variant="destructive">
        <AlertCircle className="h-4 w-4" />
        <AlertTitle>Error</AlertTitle>
        <AlertDescription>
          Failed to load challenges. Please try again later.
        </AlertDescription>
      </Alert>
    );
  }

  return (
    <div className="container mx-auto py-8">
      <h1 className="text-3xl font-bold mb-8 flex items-center gap-2">
        Coding Challenges
      </h1>
      {loading
        ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {[...Array(6)].map((_, index) => (
              <Card key={index} className="h-[250px]">
                <CardHeader>
                  <Skeleton className="h-6 w-3/4" />
                  <Skeleton className="h-4 w-1/2" />
                </CardHeader>
                <CardContent>
                  <Skeleton className="h-4 w-full mb-2" />
                  <Skeleton className="h-4 w-5/6" />
                </CardContent>
              </Card>
            ))}
          </div>
        )
        : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {data?.challenges.map((challenge) => (
              <ChallengeCard key={challenge.folderName} challenge={challenge} />
            ))}
          </div>
        )}
    </div>
  );
}
