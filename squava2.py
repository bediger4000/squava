#!/usr/bin/env python2

from math import *
import random
import string

class GameState:
    """ A state of the game, i.e. the game board. These are the only functions which are
        absolutely necessary to implement UCT in any 2-player complete information deterministic 
        zero-sum game, although they can be enhanced and made quicker, for example by using a 
        GetRandomMove() function to generate a random move during rollout.
        By convention the players are numbered 1 and 2.
    """
    def __init__(self):
            self.playerJustMoved = 2 # At the root pretend the player just moved is player 2 - player 1 has the first move
        
    def Clone(self):
        """ Create a deep clone of this game state.
        """
        st = GameState()
        st.playerJustMoved = self.playerJustMoved
        return st

    def DoMove(self, move):
        """ Update a state by carrying out the given move.
            Must update playerJustMoved.
        """
        self.playerJustMoved = 0 - self.playerJustMoved
        
    def GetMoves(self):
        """ Get all possible moves from this state.
        """
    
    def GetResult(self, playerjm):
        """ Get the game result from the viewpoint of playerjm. 
        """

    def __repr__(self):
        """ Don't need this - but good style.
        """
        pass


class SquavaState:
    """ A state of the game, i.e. the game board.
        Squares in the board are in this arrangement
 0  1  2  3  4  
 5  6  7  8  9  
10 11 12 13 14 
15 16 17 18 19 
20 21 22 23 24 
        where 0 = empty, 1 = player 1 (X), -1 = player 2 (O)
    """
    # Don't have to check the empty winning quads
    winning_quads   = [
        [],[],
        [[0,1,2,3],[1,2,3,4],[2,7,12,17]],
        [],[],[],[],
        [[5,6,7,8],[6,7,8,9],[7,12,17,22]],
        [],[],
        [[0,5,10,15],[5,10,15,20]],
        [[1,6,11,16],[6,11,16,21],[3,7,11,15],[5,11,17,23]],
        [[10,11,12,13],[11,12,13,14],[4,8,12,16],[8,12,16,20],[0,6,12,18],[6,12,18,24]],
        [[3,8,13,18],[8,13,18,23],[1,7,13,19],[9,13,17,21]],
        [[4,9,14,19],[9,14,19,24]],
        [],[],
        [[15,16,17,18],[16,17,18,19]],
        [],[],[],[],
        [[20,21,22,23],[21,22,23,24]],
        [],[]
    ]
    losing_triplets = [
        [], [],
        [ [0, 1, 2], [1, 2, 3], [2, 3, 4], [2, 7, 12], [2, 6, 10], [14, 8, 2], ],
        [], [], [], [],
        [ [5, 6, 7], [6, 7, 8], [7, 8, 9], [2, 7, 12], [7, 12, 17], [3, 7, 11], [7, 11, 15], [1, 7, 13], [7, 13, 19], ],
        [], [],
        [ [10, 11, 12], [0, 5, 10], [5, 10, 15], [10, 15, 20], [2, 6, 10], [10, 16, 22], ],
        [ [10, 11, 12], [11, 12, 13], [1, 6, 11], [6, 11, 16], [11, 16, 21], [3, 7, 11], [7, 11, 15], [5, 11, 17], [11, 17, 23], ],
        [ [10, 11, 12], [11, 12, 13], [12, 13, 14], [2, 7, 12], [7, 12, 17], [12, 17, 22], [0, 6, 12], [6, 12, 18], [12, 18, 24], [4, 8, 12], [8, 12, 16], [12, 16, 20], ],
        [ [11, 12, 13], [12, 13, 14], [3, 8, 13], [8, 13, 18], [13, 18, 23], [1, 7, 13], [7, 13, 19], [21, 17, 13], [17, 13, 9], ],
        [ [12, 13, 14], [4, 9, 14], [9, 14, 19], [14, 19, 24], [22, 18, 14], [14, 8, 2], ],
        [], [],
        [ [15, 16, 17], [16, 17, 18], [17, 18, 19], [7, 12, 17], [12, 17, 22], [5, 11, 17], [11, 17, 23], [21, 17, 13], [17, 13, 9], ],
        [], [], [], [],
        [ [20, 21, 22], [21, 22, 23], [22, 23, 24], [12, 17, 22], [10, 16, 22], [22, 18, 14], ],
        [], []
    ]
    def __init__(self):
        self.playerJustMoved = -1 # At the root pretend the player just moved is p2 - p1 has the first move
        self.board = [0 for x in range(25)] # 0 = empty, 1 = player 1, 2 = player 2
        
    def Clone(self):
        """ Create a deep clone of this game state.
        """
        st = SquavaState()
        st.playerJustMoved = self.playerJustMoved
        st.board = self.board[:]
        return st

    def DoMove(self, move):
        self.playerJustMoved = 0 - self.playerJustMoved
        self.board[move] = self.playerJustMoved

        
    def GetMoves(self):
        moves = []
        for i in range(25):
            if self.board[i] == 0:
                moves.append(i)
            else:
                row = SquavaState.winning_quads[i]
                if len(row):
                    for quad in row:
                        sum = self.board[quad[0]] + self.board[quad[1]] + self.board[quad[2]] + self.board[quad[3]]
                        if sum == 4 or sum == -4:
                            return []
                row = SquavaState.losing_triplets[i]
                if len(row):
                    for triplet in row:
                        sum = self.board[triplet[0]] + self.board[triplet[1]] + self.board[triplet[2]]
                        if sum == 3 or sum == -3:
                            return []
        return moves
    
    def GetResult(self, playerjm):
        for row in SquavaState.winning_quads:
            if len(row):
                for quad in row:
                    sum = self.board[quad[0]] + self.board[quad[1]] + self.board[quad[2]] + self.board[quad[3]]
                    if sum == 4 or sum == -4:
                        if sum == 4*playerjm:
                            return 1.0
                        else:
                            return 0.0

        for row in SquavaState.losing_triplets:
            if len(row):
                for triplet in row:
                    sum = self.board[triplet[0]] + self.board[triplet[1]] + self.board[triplet[2]]
                    if sum == 3 or sum == -3:
                        if sum == 3*playerjm:
                            return 0.0
                        else:
                            return 1.0
        print "This should not happen:"
        print str(self)
        assert False
        return 0.0 # I think it never gets here

    def __repr__(self):
        s= "   0 1 2 3 4\n"
        for i in range(25): 
            if i % 5 == 0: s += str(i/5)+'  '
            s += "_XO"[self.board[i]]
            s += ' '
            if i % 5 == 4: s += "\n"
        return s

class Node:
    def __init__(self, move = None, parent = None, state = None):
        self.move = move
        self.parentNode = parent # "None" for the root node
        self.childNodes = []
        self.wins = 0
        self.visits = 0
        self.untriedMoves = state.GetMoves() # future child nodes
        self.playerJustMoved = state.playerJustMoved # the only part of the state that the Node needs later
        
    def UCTSelectChild(self):
        """ Use the UCB1 formula to select a child node. Often a constant
        UCTK is applied so we have lambda c: c.wins/c.visits + UCTK *
        sqrt(2*log(self.visits)/c.visits to vary the amount of exploration
        versus exploitation.
        """
        s = sorted(self.childNodes, key = lambda c: c.wins/c.visits + sqrt(2*log(self.visits)/c.visits))[-1]
        return s
    
    def AddChild(self, m, s):
        """ Remove m from untriedMoves and add a new child node for this move.
            Return the added child node
        """
        n = Node(move = m, parent = self, state = s)
        self.untriedMoves.remove(m)
        self.childNodes.append(n)
        return n
    
    def Update(self, result):
        """ Update this node - one additional visit and result additional wins. result must be from the viewpoint of playerJustmoved.
        """
        self.visits += 1
        self.wins += result

    def __repr__(self):
        return "[M:" + str(self.move) + " W/V:" + str(self.wins) + "/" + str(self.visits) + " U:" + str(self.untriedMoves) + "]"

    def TreeToString(self, indent):
        s = self.IndentString(indent) + str(self)
        for c in self.childNodes:
             s += c.TreeToString(indent+1)
        return s

    def IndentString(self,indent):
        s = "\n"
        for i in range (1,indent+1):
            s += "| "
        return s

    def ChildrenToString(self):
        s = ""
        for c in self.childNodes:
             s += str(c) + "\n"
        return s


def UCT(rootstate, itermax, verbose = False):
    """ Conduct a UCT search for itermax iterations starting from rootstate.
        Return the best move from the rootstate.
        Assumes 2 alternating players (player 1 starts), with game results in the range [0.0, 1.0]."""

    rootnode = Node(state = rootstate)

    for i in range(itermax):

        node = rootnode
        state = rootstate.Clone()

        # Select
        while node.untriedMoves == [] and node.childNodes != []: # node is fully expanded and non-terminal
            node = node.UCTSelectChild()
            state.DoMove(node.move)

        # Expand
        if node.untriedMoves != []: # if we can expand (i.e. state/node is non-terminal)
            m = random.choice(node.untriedMoves) 
            state.DoMove(m)
            node = node.AddChild(m,state) # add child and descend tree

        # Rollout - this can often be made orders of magnitude quicker
        # using a state.GetRandomMove() function
        while state.GetMoves() != []: # while state is non-terminal
            state.DoMove(random.choice(state.GetMoves()))

        # Backpropagate
        while node != None: # backpropagate from the expanded node and work back to the root node
            r = state.GetResult(node.playerJustMoved) # state is terminal. Update node with result from POV of node.playerJustMoved
            print "Updating node for move",node.move,", player ", node.playerJustMoved, ", result ", r
            node.Update(r)
            node = node.parentNode

    # Output some information about the tree - can be omitted
    #if (verbose): print rootnode.TreeToString(0)
    #else: print rootnode.ChildrenToString()

    return sorted(rootnode.childNodes, key = lambda c: c.visits)[-1].move # return the move that was most visited
                
def UCTPlayGame():
    """ Play a sample game between two UCT players where each player gets
        a different number of UCT iterations (= simulations = tree nodes).
    """
    state = SquavaState()
    while (state.GetMoves() != []):
        print str(state)
        if state.playerJustMoved == 1:
            m = UCT(rootstate = state, itermax = 25, verbose = False)
        else:
            a = string.split(raw_input('Enter move (N M): '))
            m = int(a[0])*5 + int(a[1])
        state.DoMove(m)
        print 'Player '+ "_XO"[state.playerJustMoved] + " chose: " + str(m) + "\n"
    print str(state)
    if state.GetResult(state.playerJustMoved) == 1.0:
        print "Player " + "_XO"[state.playerJustMoved] + " wins!"
    elif state.GetResult(state.playerJustMoved) == 0.0:
        print "Player " + "_XO"[state.playerJustMoved] + " loses!"
    else: print "Nobody wins!"

if __name__ == "__main__":
    """ Play a single game to the end using UCT for both players. 
    """
    UCTPlayGame()

            
                          
            

