import { NextResponse } from 'next/server';

export const dynamic = 'force-dynamic';
export const maxDuration = 120;

const backendUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:3001';

export async function POST() {
  try {
    const response = await fetch(`${backendUrl}/api/demo/load-data`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: '{}',
      cache: 'no-store',
    });

    const text = await response.text();
    const contentType = response.headers.get('content-type') || 'application/json';

    return new NextResponse(text, {
      status: response.status,
      headers: { 'Content-Type': contentType },
    });
  } catch (error) {
    const message = error instanceof Error ? error.message : 'unknown error';
    return NextResponse.json(
      { error: `Failed to seed demo data: ${message}` },
      { status: 502 },
    );
  }
}
