import axios, { AxiosError, CancelTokenSource } from "axios";
import { useEffect, useState } from "react";

interface GqlResponse<T> {
  data: T;
  errors?: Array<{ message: string }>;
}

export const GRAPHQL_ENDPOINT = "/api/query";

function useGql<T>(query: string) {
  const [data, setData] = useState<T | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    // Cancel token to handle component unmount
    const source: CancelTokenSource = axios.CancelToken.source();
    const fetchData = async () => {
      setLoading(true);
      setError(null);

      try {
        const response = await axios({
          method: "POST",
          url: GRAPHQL_ENDPOINT,
          data: {
            query,
          },
          headers: {
            "Content-Type": "application/json",
          },
          cancelToken: source.token,
        });

        const result = response.data as GqlResponse<T>;
        if (result.errors && result.errors.length > 0) {
          throw new Error(result.errors[0].message);
        }

        setData(result.data);
      } catch (err) {
        if (axios.isCancel(err)) {
          // console.log('Request canceled:', err.message)
        } else {
          const message =
            err instanceof AxiosError && err.response?.data?.errors
              ? err.response.data.errors[0].message
              : "An unexpected error occurred";
          setError(new Error(message));
        }
      } finally {
        setLoading(false);
      }
    };

    fetchData();

    // Cleanup on unmount
    return () => {
      source.cancel("Request canceled due to component unmount");
    };
  }, [query]);

  return { data, loading, error };
}

export default useGql;
