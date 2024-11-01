import axios from "axios";
import { useEffect, useState } from "react";

function useFetch<T>(
  url: string,
  query?: Record<string, any>,
  method: "GET" | "POST" = "GET",
) {
  const [data, setData] = useState<T | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const response = await axios({
          method,
          url,
          params: query,
          headers: {
            "Content-Type": "application/json",
          },
        });
        setData(response.data);
      } catch (err) {
        setError(err instanceof Error ? err : new Error("An error occurred"));
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [url, query, method]);

  return { data, loading, error };
}

interface ResponseWrapper<T> {
  timestamp: string;
  isError: boolean;
  errorMessage: string;
  data: T;
}

export { type ResponseWrapper, useFetch };
