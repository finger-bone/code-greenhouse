import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Progress } from "@/components/ui/progress";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import useGql from "@/hooks/useGql";
import { useTheme } from "@/providers/theme-provider";
import axios from "axios";
import { AnimatePresence, motion } from "framer-motion";
import {
  Check,
  ChevronLeft,
  ChevronRight,
  Circle,
  Clock,
  GitFork,
  Library,
  Play,
  X,
} from "lucide-react";
import { useEffect, useState } from "react";
import { useParams } from "react-router-dom";

// Type Definitions
interface RepositoryData {
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

interface StageData {
  name: string;
  description: string;
  noteFileOrPath: string;
  noteFileType: string;
}

interface ChallengeData {
  folderName: string;
  stages: StageData[];
}

interface SidebarProps {
  isSidebarOpen: boolean;
  setIsSidebarOpen: React.Dispatch<React.SetStateAction<boolean>>;
  repoData: { repository: RepositoryData } | null;
  challengeData: { challenge: ChallengeData } | null;
  currentStage: number;
  setCurrentStage: React.Dispatch<React.SetStateAction<number>>;
}

interface MainContentProps {
  isSidebarOpen: boolean;
  setIsSidebarOpen: React.Dispatch<React.SetStateAction<boolean>>;
  challengeData: { challenge: ChallengeData } | null;
  currentStage: number;
  theme: string;
}

interface FooterProps {
  currentStage: number;
  repoData: { repository: RepositoryData } | null;
  handlePrevStage: () => void;
  handleNextStage: () => void;
  handleRequestTest: () => void;
  handleCopyGitUrl: () => void;
  handleToggleTestingPanel: () => void;
}

interface DialogState {
  isOpen: boolean;
  title: string;
  description: string;
}


interface TestingData {
  serial: number;
  status: string;
  message: string;
  createTime: string;
  runStartTime: string;
  runEndTime: string;
}

export default function RepositoryPage() {
  const { repoId } = useParams<{ repoId: string }>();
  const [isSidebarOpen, setIsSidebarOpen] = useState<boolean>(true);
  const [currentStage, setCurrentStage] = useState<number>(1);
  const [dialogState, setDialogState] = useState<DialogState>({
    isOpen: false,
    title: "",
    description: "",
  });
  const theme = useTheme().theme;

  const [testingsPanelOpen, setTestingsPanelOpen] = useState<boolean>(false);
  const [testings, setTestings] = useState<TestingData[]>([]);

  const {
    data: repoData,
    loading: repoLoading,
    error: repoError,
  } = fetchRepoData(repoId);
  const {
    data: challengeData,
    loading: challengeLoading,
    error: challengeError,
  } = fetchChallengeData(repoData?.repository.challengeFolderName);

  useEffect(() => {
    setCurrentStage(repoData?.repository.stage || 1);
  }, [repoData]);

  useEffect(() => {
    if (testingsPanelOpen) {
      fetchTestings(repoId!, currentStage).then((testings) => {
        setTestings(testings);
      });

      const intervalId = setInterval(() => {
        fetchTestings(repoId!, currentStage).then((testings) => {
          setTestings(testings);
        });
      }, 5000); // Fetch every 5 seconds

      return () => clearInterval(intervalId);
    }
  }, [repoId, currentStage, testingsPanelOpen]);

  if (repoLoading || challengeLoading) return <LoadingScreen />;
  if (repoError || challengeError) return <ErrorScreen />;

  return (
    <div className="flex flex-col overflow-hidden w-full h-[calc(100vh-72px)]">
      <div className="flex flex-grow overflow-hidden">
        <Sidebar
          isSidebarOpen={isSidebarOpen}
          setIsSidebarOpen={setIsSidebarOpen}
          repoData={repoData}
          challengeData={challengeData}
          currentStage={currentStage}
          setCurrentStage={setCurrentStage}
        />
        <MainContent
          isSidebarOpen={isSidebarOpen}
          setIsSidebarOpen={setIsSidebarOpen}
          challengeData={challengeData}
          currentStage={currentStage}
          theme={theme}
        />
        <TestingPanel
          isOpen={testingsPanelOpen}
          onClose={() => setTestingsPanelOpen(false)}
          testings={testings}
        />
      </div>
      <Footer
        currentStage={currentStage}
        repoData={repoData}
        handlePrevStage={() => setCurrentStage(currentStage - 1)}
        handleNextStage={() => setCurrentStage(currentStage + 1)}
        handleRequestTest={() =>
          handleRequestTest(repoId, currentStage, setDialogState)
        }
        handleCopyGitUrl={() => handleCopyGitUrl(repoData, setDialogState)}
        handleToggleTestingPanel={() => setTestingsPanelOpen(!testingsPanelOpen)}
      />
      <Dialog
        open={dialogState.isOpen}
        onOpenChange={(isOpen) =>
          setDialogState((prev) => ({ ...prev, isOpen }))
        }
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{dialogState.title}</DialogTitle>
            <DialogDescription>{dialogState.description}</DialogDescription>
          </DialogHeader>
        </DialogContent>
      </Dialog>
    </div>
  );
}

// Helper Functions

function fetchRepoData(repoId: string | undefined) {
  return useGql<{ repository: RepositoryData }>(`
    query Repository {
      repository(repositoryId: "${repoId}") {
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
}

function fetchChallengeData(challengeFolderName: string | undefined) {
  return useGql<{ challenge: ChallengeData }>(`
    query Challenge {
      challenge(folderName: "${challengeFolderName}") {
        folderName
        stages {
          name
          description
          noteFileOrPath
          noteFileType
        }
      }
    }
  `);
}

import { GRAPHQL_ENDPOINT } from "@/hooks/useGql";
import { use } from "i18next";

async function fetchTestings(
  repoId: string,
  stage: number
): Promise<
  {
    serial: number;
    status: string;
    message: string;
    createTime: string;
    runStartTime: string;
    runEndTime: string;
  }[]
> {
  const query = `
  query TestingsByStage {
    testingsByStage(repositoryId: "${repoId}", stage: ${stage}) {
        serial
        status
        message
        createTime
        runStartTime
        runEndTime
    }
  }
  `;
    const respData = (await axios({
      method: "POST",
      url: GRAPHQL_ENDPOINT,
      data: {
        query,
      },
      headers: {
        "Content-Type": "application/json",
      },
    })).data as {
      data: {
        testingsByStage: {
          serial: number;
          status: string;
          message: string;
          createTime: string;
          runStartTime: string;
          runEndTime: string;
        }[];
      }
    }
    return respData.data.testingsByStage
}

function LoadingScreen() {
  return (
    <div className="flex flex-col h-screen items-center justify-center">
      <Skeleton className="h-12 w-12 rounded-full" />
      <Skeleton className="h-4 w-[200px] mt-4" />
    </div>
  );
}

function ErrorScreen() {
  return (
    <div className="flex flex-col h-screen items-center justify-center text-destructive">
      <p className="text-xl font-semibold">Error loading data</p>
      <p className="mt-2">Please try again later</p>
    </div>
  );
}

function Sidebar({
  isSidebarOpen,
  setIsSidebarOpen,
  repoData,
  challengeData,
  currentStage,
  setCurrentStage,
}: SidebarProps) {
  return (
    <AnimatePresence>
      {isSidebarOpen && (
        <motion.aside
          initial={{ width: 0, opacity: 0 }}
          animate={{ width: "25%", opacity: 1 }}
          exit={{ width: 0, opacity: 0 }}
          transition={{ duration: 0.3 }}
          className="bg-secondary flex flex-col"
        >
          <div className="p-4 border-b">
            <h2 className="text-xl font-bold mb-2">Stages</h2>
            <Progress
              value={
                (((repoData?.repository.stage ?? 0) /
                  (repoData?.repository.totalStages ?? 1)) *
                  100 *
                  99) /
                  100 +
                1
              }
              className="h-2"
            />
            <p className="text-sm text-muted-foreground mt-1">
              {repoData?.repository.stage} stages finished out of{" "}
              {repoData?.repository.totalStages}
            </p>
          </div>
          <ScrollArea className="flex-grow">
            <nav className="p-2">
              {challengeData?.challenge.stages.map((stage, index) => (
                <Button
                  key={index}
                  variant={currentStage === index + 1 ? "secondary" : "ghost"}
                  className="w-full justify-start mb-1 text-left"
                  onClick={() => setCurrentStage(index + 1)}
                >
                  {getStageIcon(index, repoData?.repository.stage!)}
                  <span className="ml-2">{index + 1}.</span> {stage.name}
                </Button>
              ))}
            </nav>
          </ScrollArea>
        </motion.aside>
      )}
    </AnimatePresence>
  );
}

function MainContent({
  isSidebarOpen,
  setIsSidebarOpen,
  challengeData,
  currentStage,
  theme,
}: MainContentProps) {
  return (
    <main className="flex flex-col overflow-hidden w-full">
      <header className="flex items-center p-4 bg-background border-b">
        <Button
          variant="ghost"
          size="icon"
          onClick={() => setIsSidebarOpen(!isSidebarOpen)}
          aria-label={isSidebarOpen ? "Close sidebar" : "Open sidebar"}
        >
          {isSidebarOpen ? <ChevronLeft /> : <ChevronRight />}
        </Button>
        <h1 className="text-2xl font-bold ml-4">
          Stage {currentStage}:{" "}
          {challengeData?.challenge.stages[currentStage - 1].name}
        </h1>
      </header>
      <div className="flex-grow overflow-auto p-6">
        <iframe
          className="w-full h-full border rounded-lg"
          src={`/api/note/${challengeData?.challenge.folderName}/${
            currentStage - 1
          }/?theme=${theme}`}
          title={`Stage ${currentStage} content`}
        />
      </div>
    </main>
  );
}


function TestingPanel({ isOpen, onClose, testings }: { isOpen: boolean; onClose: () => void; testings: TestingData[] }) {
  return (
    <AnimatePresence>
      {isOpen && (
        <motion.div
          initial={{ width: 0, opacity: 0 }}
          animate={{ width: "30%", opacity: 1 }}
          exit={{ width: 0, opacity: 0 }}
          transition={{ duration: 0.3 }}
          className="bg-background border-l overflow-hidden"
        >
          <div className="flex justify-between items-center p-4 border-b">
            <h2 className="text-xl font-bold">Testing Information</h2>
            <Button variant="ghost" size="icon" onClick={onClose}>
              <X className="h-4 w-4" />
            </Button>
          </div>
          <ScrollArea className="h-full">
            <div className="p-4">
              {testings.map((testing) => (
                <div key={testing.serial} className="mb-4 p-4 border rounded-lg">
                  <div className="flex justify-between items-center mb-2">
                    <span className="font-semibold">Test #{testing.serial}</span>
                    <span className={`px-2 py-1 rounded-full text-xs ${
                      testing.status === 'SUCCESS' ? 'bg-green-100 text-green-800' :
                      testing.status === 'FAILURE' ? 'bg-red-100 text-red-800' :
                      'bg-yellow-100 text-yellow-800'
                    }`}>
                      {testing.status}
                    </span>
                  </div>
                  <p className="text-sm mb-2">{testing.message}</p>
                  <div className="text-xs text-muted-foreground">
                    <p>Created: {new Date(testing.createTime).toLocaleString()}</p>
                    <p>Started: {testing.runStartTime ? new Date(testing.runStartTime).toLocaleString() : 'N/A'}</p>
                    <p>Ended: {testing.runEndTime ? new Date(testing.runEndTime).toLocaleString() : 'N/A'}</p>
                  </div>
                </div>
              ))}
            </div>
          </ScrollArea>
        </motion.div>
      )}
    </AnimatePresence>
  );
}

function Footer({
  currentStage,
  repoData,
  handlePrevStage,
  handleNextStage,
  handleRequestTest,
  handleCopyGitUrl,
  handleToggleTestingPanel,
}: FooterProps & { handleToggleTestingPanel: () => void }) {
  return (
    <TooltipProvider>
      <footer className="flex justify-between items-center p-4 bg-muted border-t">
        <Button onClick={handlePrevStage} disabled={currentStage === 1}>
          <ChevronLeft className="mr-2 h-4 w-4" /> Previous
        </Button>
        <div>
          <Button onClick={handleRequestTest} variant="ghost">
            <Play className="mr-2 h-4 w-4" /> Test Run For This Stage
          </Button>
          <Tooltip>
            <TooltipTrigger asChild>
              <Button onClick={handleCopyGitUrl} variant="ghost">
                <GitFork className="mr-2 h-4 w-4" /> Copy Git URL
              </Button>
            </TooltipTrigger>
            <TooltipContent>
              <p>Copy Git URL to clipboard</p>
            </TooltipContent>
          </Tooltip>
          <Button onClick={handleToggleTestingPanel} variant="ghost">
            <Library className="mr-2 h-4 w-4" /> View Testing Info
          </Button>
        </div>
        <Button
          onClick={handleNextStage}
          disabled={currentStage === repoData?.repository.totalStages}
        >
          Next <ChevronRight className="ml-2 h-4 w-4" />
        </Button>
      </footer>
    </TooltipProvider>
  );
}

// Helper Function for making API calls
async function handleRequestTest(
  repoId: string | undefined,
  currentStage: number,
  setDialogState: React.Dispatch<React.SetStateAction<DialogState>>
) {
  try {
    const params = new URLSearchParams();
    params.append("repo", repoId || "");
    params.append("stage", currentStage.toString());
    const resp = await axios.post(`/api/testing/pending?${params.toString()}`);
    setDialogState({
      isOpen: true,
      title:
        resp.status === 200
          ? "Successfully requested test"
          : "Test already requested",
      description:
        resp.status === 200
          ? "The test is in the queue."
          : "The test has already been requested.",
    });
  } catch {
    setDialogState({
      isOpen: true,
      title: "Error",
      description: "Could not request test.",
    });
  }
}

function handleCopyGitUrl(
  repoData: { repository: RepositoryData } | null,
  setDialogState: React.Dispatch<React.SetStateAction<DialogState>>
) {
  if (!repoData?.repository.startpoint) return;
  const gitUrl = `${window.location.origin}/api/repo/git/${repoData?.repository.provider}/${repoData?.repository.subject}/${repoData.repository.challengeFolderName}/${repoData?.repository.repositoryId}`;
  navigator.clipboard.writeText(gitUrl);
  setDialogState({
    isOpen: true,
    title: "Git URL copied",
    description: "The URL has been copied to your clipboard.",
  });
}

function getStageIcon(stageIndex: number, currentStage: number) {
  if (stageIndex < currentStage) {
    return <Check className="mr-2 h-4 w-4 text-emerald-500" />;
  }
  if (stageIndex === currentStage) {
    return <Circle className="mr-2 h-4 w-4 text-blue-500" />;
  }
  return <Clock className="mr-2 h-4 w-4 text-gray-500" />;
}
