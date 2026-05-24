import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

export function middleware(request: NextRequest) {
  const token = request.cookies.get('oa-auth')?.value;
  const { pathname } = request.nextUrl;

  const protectedPaths = ['/app'];
  //const publicPaths = ['/login', '/', '/demo', '/mission', '/pricing', '/performance'];

  const isProtected = protectedPaths.some(path => pathname.startsWith(path));
  //const isPublic = publicPaths.some(path => pathname.startsWith(path));

  if (!token && isProtected) {
    const url = request.nextUrl.clone();
    url.pathname = '/';
    return NextResponse.redirect(url);
  }

  if (token && pathname === '/login') {
    const url = request.nextUrl.clone();
    url.pathname = '/app/dashboard';
    return NextResponse.redirect(url);
  }

  return NextResponse.next();
}

export const config = {
  matcher: ['/((?!_next/api|_next/static|favicon.ico|public).*)', ],
};