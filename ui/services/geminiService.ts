
import { GoogleGenAI, Type } from "@google/genai";
import { AuditLog, ApiToken } from "../types";

const ai = new GoogleGenAI({ apiKey: process.env.API_KEY || '' });

export const analyzeSecurityLogs = async (logs: AuditLog[]) => {
  const logSummary = logs.map(l => `${l.timestamp}: ${l.method} ${l.path} - ${l.status}`).join('\n');
  
  const response = await ai.models.generateContent({
    model: 'gemini-3-flash-preview',
    contents: `Analyze these Grafana Gateway proxy logs for security threats, rate-limit abuse, or unusual patterns: \n\n${logSummary}`,
    config: {
      systemInstruction: "You are a senior security engineer. Provide a concise JSON response summarizing threats.",
      responseMimeType: "application/json",
      responseSchema: {
        type: Type.OBJECT,
        properties: {
          threatLevel: { type: Type.STRING, description: 'LOW, MEDIUM, HIGH' },
          summary: { type: Type.STRING },
          anomalies: {
            type: Type.ARRAY,
            items: { type: Type.STRING }
          },
          recommendations: {
            type: Type.ARRAY,
            items: { type: Type.STRING }
          }
        },
        required: ['threatLevel', 'summary', 'anomalies', 'recommendations']
      }
    }
  });

  return JSON.parse(response.text);
};

export const suggestRateLimit = async (teamName: string, useCase: string) => {
  const response = await ai.models.generateContent({
    model: 'gemini-3-flash-preview',
    contents: `Suggest a sensible API rate limit (requests per minute) and RBAC policy for a team named "${teamName}" using Grafana for "${useCase}".`,
    config: {
      systemInstruction: "Provide professional infrastructure advice.",
      responseMimeType: "application/json",
      responseSchema: {
        type: Type.OBJECT,
        properties: {
          recommendedRPM: { type: Type.NUMBER },
          reasoning: { type: Type.STRING },
          suggestedDashboardPatterns: { type: Type.ARRAY, items: { type: Type.STRING } }
        },
        required: ['recommendedRPM', 'reasoning', 'suggestedDashboardPatterns']
      }
    }
  });

  return JSON.parse(response.text);
};
