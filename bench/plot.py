import matplotlib.pyplot as plt



x = [100, 
200, 
400, 
500, 
1000, 
2000]


z_for_2 = [98272898488,
92205496602,
84612637888,
79918565643,
79061862865,
76362242882]

z_for_2 = [(10**9)*2000/i for i in z_for_2]


z_for_4 = [99664095212,
74156690763,
65680760223,
63432080879,
62776608150,
58708567933]

z_for_4 = [(10**9)*2000/i for i in z_for_4]

z_for_8 = [89466870308,
58968211836,
47211272890,
45488409001,
42521193888,
42058535816]

z_for_8 = [(10**9)*2000/i for i in z_for_8]



import plotly.express as px
import plotly.graph_objects as go


fig = go.Figure()

# Add traces
fig.add_trace(go.Scatter(x=x, y=z_for_8,
                    mode='lines',
                    name='8 ядер'))
# fig.add_trace(go.Scatter(x=x1, y=y1,
#                     mode='lines',
#                     name='2 ядра'))

fig.add_trace(go.Scatter(x=x, y=z_for_4,
                    mode='lines',
                    name='4 ядра'))

fig.add_trace(go.Scatter(x=x, y=z_for_2,
                    mode='lines',
                    name='2 ядра'))

# fig.add_mar
# fig.add_tex

fig.update_layout(
    title="",
    xaxis_title="Кількість горутин(зелених тредів)",
    yaxis_title="Швидкість парсингу(домен в секунду)",
    legend_title="",
    font=dict(
        family="Courier New, monospace",
        size=18,
        color="RebeccaPurple"
    )
)

fig.show()