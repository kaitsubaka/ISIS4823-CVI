import pygame as pg
from pygame.locals import *
from OpenGL.GL import *
from OpenGL.GLU import *
from math import * 

verticies = (
    (0.25, -0.25, -1),
    (0.25, 0.25, -1),
    (-0.25, 0.25, -1),
    (-0.25, -0.25, -1),
    (0.25, -0.25, 1),
    (0.25, 0.25, 0.9),
    (-0.25, -0.25, 1),
    (-0.25, 0.25, 0.9),
    (-0.25, -0.10, 1),
    (0.25, -0.10, 1),
    )

edges = (
    (0,1),
    (0,3),
    (0,4),
    (4,9),
    (2,1),
    (2,3),
    (2,7),
    (6,3),
    (6,4),
    (6,8),
    (8,7),
    (5,1),
    (5,9),
    (5,7)


    )

widowVertex = (
    (0.2, 0.2, 0.92),
    (0.2, -0.10, 1),
    (-0.2, -0.10, 1),
    (-0.2, 0.2, 0.92)
)

windowEdges = (
    (0,1),
    (1,2),
    (2,3),
    (3,0)
)


def cicle_skeleton(r,n=100):
    return [(cos(2*pi/n*x)*r,sin(2*pi/n*x)*r) for x in range(0,n)]

n=10
circle = cicle_skeleton(0.1,n)
wheelVerticies= ()
wheelEdges = ()
k = i = len(wheelVerticies)


for j in range(8):
    while i<k+n-1:
        wheelEdges+=((i,i+1),)
        if j % 2 == 0:
            wheelEdges+=((i,i+n),)
        i+=1
    
    if j % 2 == 0:
            wheelEdges+=((i,i+n),)
    wheelEdges+=((k,i),)
    k+=n
    i=k


wheelVerticies+=tuple(map(lambda x: (0.3,x[0]-0.25,x[1]-0.7), circle))
wheelVerticies+=tuple(map(lambda x: (0.2,x[0]-0.25,x[1]-0.7), circle))
wheelVerticies+=tuple(map(lambda x: (-0.3,x[0]-0.25,x[1]-0.7), circle))
wheelVerticies+=tuple(map(lambda x: (-0.2,x[0]-0.25,x[1]-0.7), circle))

wheelVerticies+=tuple(map(lambda x: (0.3,x[0]-0.25,x[1]+0.7), circle))
wheelVerticies+=tuple(map(lambda x: (0.2,x[0]-0.25,x[1]+0.7), circle))
wheelVerticies+=tuple(map(lambda x: (-0.3,x[0]-0.25,x[1]+0.7), circle))
wheelVerticies+=tuple(map(lambda x: (-0.2,x[0]-0.25,x[1]+0.7), circle))

def draw_cube():
    glBegin(GL_LINES)
    glColor3f(1.0,0.0,0.0)
    for edge in edges:
        for vertex in edge:
            glVertex3fv(verticies[vertex])
    glEnd()

def draw_wheels():
    glBegin(GL_LINES)
    glColor3f(0.0,0.0,0.0)
    for edge in wheelEdges:
        for vertex in edge:
            glVertex3fv(wheelVerticies[vertex])
    glEnd()
    
    
def draw_widow():
    glBegin(GL_LINES)
    glColor3f(0.0,0.0,1.0)
    for edge in windowEdges:
        for vertex in edge:
            glVertex3fv(widowVertex[vertex])
    glEnd()

def main():
    pg.init()
    
    display = (800,600)
    
    pg.display.set_mode(display, DOUBLEBUF|OPENGL)
    
    gluPerspective(45, (display[0]/display[1]), 0.1, 50.0)
    
    glTranslatef(0.0,0.0, -5)
    
    
    
    while True:
        for event in pg.event.get():
            if event.type == pg.QUIT:
                pg.quit()
                quit()
        glRotatef(1, 0, 1, 0)
        
        glClear(GL_COLOR_BUFFER_BIT|GL_DEPTH_BUFFER_BIT)
        glClearColor(255,255,255,255)
        draw_cube()
        draw_wheels()
        draw_widow()
        pg.display.flip()
        pg.time.wait(10)

if __name__ == '__main__':
    main()
    