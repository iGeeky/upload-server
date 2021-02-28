
#
CC=gcc
all: libxfile.$(shell uname).a
CFLAGS=-g -D_FILE_OFFSET_BITS=64 -D_LARGE_FILE -Werror -fPIC -I c
LDFLAGS=-lcrypto

SRCS=xfile.c

CFLAGS_A=-g -D_FILE_OFFSET_BITS=64 -D_LARGE_FILE -Werror

libxfile.so: ${SRCS}
	$(CC) ${CFLAGS} ${LDFLAGS} -Werror $^ -fPIC -shared -o $@

libxfile.$(shell uname).a: ${SRCS}
	$(CC) ${CFLAGS_A} -c $^
	ar rc $@ *.o
clean:
	rm -f *.so *.a  *.o *.oxx
