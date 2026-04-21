import { useState, useCallback } from "react";
import { useMutation } from "@tanstack/react-query";
import { examSubmissionService } from "@/services/exam-submission.service";
import * as types from "@/types/exam-submission.types";
import toast from "react-hot-toast";

interface SubmissionState {
  submittedProblems: Map<number, types.ProblemSubmissionResult>;
  isSubmitting: boolean;
  currentSubmission: types.ProblemSubmissionResult | null;
  error: string | null;
}

/**
 * Hook to manage exam problem submissions
 */
export const useExamSubmission = (examID: number) => {
  const [state, setState] = useState<SubmissionState>({
    submittedProblems: new Map(),
    isSubmitting: false,
    currentSubmission: null,
    error: null,
  });

  // Mutation for submitting problem code
  const submitMutation = useMutation({
    mutationFn: async (params: {
      problemID: number;
      code: string;
      language?: string;
    }) => {
      setState((prev) => ({ ...prev, isSubmitting: true, error: null }));

      try {
        const result = await examSubmissionService.submitProblemCode(
          examID,
          params.problemID,
          {
            code: params.code,
            databaseType: params.language || 'postgresql',
          }
        );

        return result;
      } catch (error) {
        const errorMsg = error instanceof Error ? error.message : "Submission failed";
        setState((prev) => ({ ...prev, error: errorMsg }));
        throw error;
      }
    },
    onSuccess: (data) => {
      setState((prev) => {
        const newMap = new Map(prev.submittedProblems);
        newMap.set(data.examProblemID, data);

        return {
          ...prev,
          submittedProblems: newMap,
          currentSubmission: data,
          isSubmitting: false,
        };
      });

      // Show success/failure toast
      if (data.status === "accepted") {
        toast.success(`✓ Correct! +${data.score} points`);
      } else if (data.status === "wrong_answer") {
        toast.error("✗ Wrong answer, try again");
      } else {
        toast.error("Error executing code");
      }
    },
    onError: (error) => {
      setState((prev) => ({
        ...prev,
        isSubmitting: false,
      }));
      toast.error(error instanceof Error ? error.message : "Submission failed");
    },
  });

  const submit = useCallback(
    async (
      problemID: number,
      code: string,
      language: string = "postgresql"
    ) => {
      return submitMutation.mutate({
        problemID,
        code,
        language,
      });
    },
    [submitMutation]
  );

  const isProblemSolved = (problemID: number): boolean => {
    const submission = state.submittedProblems.get(problemID);
    return submission ? submission.status === "accepted" : false;
  };

  const getSubmissionAttempts = (problemID: number): number => {
    const submissions = Array.from(state.submittedProblems.values());
    return submissions.filter((s) => s.examProblemID === problemID).length;
  };

  return {
    ...state,
    submit,
    isProblemSolved,
    getSubmissionAttempts,
    submitMutation,
  };
};
